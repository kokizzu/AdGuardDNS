package cmd

import (
	"fmt"
	"log/slog"
	"net/netip"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardDNS/internal/agdcache"
	"github.com/AdguardTeam/AdGuardDNS/internal/agdhttp"
	"github.com/AdguardTeam/AdGuardDNS/internal/backendpb"
	"github.com/AdguardTeam/AdGuardDNS/internal/dnscheck"
	"github.com/AdguardTeam/AdGuardDNS/internal/dnsmsg"
	"github.com/AdguardTeam/AdGuardDNS/internal/errcoll"
	"github.com/AdguardTeam/AdGuardDNS/internal/metrics"
	"github.com/AdguardTeam/AdGuardDNS/internal/remotekv"
	"github.com/AdguardTeam/AdGuardDNS/internal/remotekv/consulkv"
	"github.com/AdguardTeam/AdGuardDNS/internal/remotekv/rediskv"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/redisutil"
	"github.com/AdguardTeam/golibs/timeutil"
	"github.com/AdguardTeam/golibs/validate"
	"github.com/c2h5oh/datasize"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

// checkConfig is the DNS server checking configuration.
type checkConfig struct {
	// KV is remote key-value store configuration for DNS server checking.
	KV *remoteKVConfig `yaml:"kv"`

	// Domains are the domain names used for DNS server checking.
	Domains []string `yaml:"domains"`

	// NodeLocation is the location of this server node.
	NodeLocation string `yaml:"node_location"`

	// NodeName is the name of this server node.
	NodeName string `yaml:"node_name"`

	// IPv4 is the list of IPv4 addresses to respond with for A queries to
	// subdomains of Domain.
	IPv4 []netip.Addr `yaml:"ipv4"`

	// IPv6 is the list of IPv6 addresses to respond with for AAAA queries to
	// subdomains of Domain.
	IPv6 []netip.Addr `yaml:"ipv6"`
}

// toInternal converts c to the DNS server check configuration for the DNS
// server.  c must be valid.
func (c *checkConfig) toInternal(
	baseLogger *slog.Logger,
	envs *environment,
	messages *dnsmsg.Constructor,
	errColl errcoll.Interface,
	namespace string,
	reg prometheus.Registerer,
	grpcMtrc backendpb.GRPCMetrics,
) (conf *dnscheck.RemoteKVConfig, err error) {
	mtrc, err := metrics.NewDNSCheck(namespace, reg)
	if err != nil {
		return nil, fmt.Errorf("dnscheck metrics: %w", err)
	}

	kv, err := c.KV.newRemoteKV(envs, namespace, reg, grpcMtrc, baseLogger)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		return nil, err
	}

	domains := make([]string, len(c.Domains))
	for i, d := range c.Domains {
		domains[i] = strings.ToLower(d)
	}

	return &dnscheck.RemoteKVConfig{
		Logger:       baseLogger.With(slogutil.KeyPrefix, "dnscheck"),
		Messages:     messages,
		Metrics:      mtrc,
		RemoteKV:     kv,
		ErrColl:      errColl,
		Domains:      domains,
		NodeLocation: c.NodeLocation,
		NodeName:     c.NodeName,
		IPv4:         c.IPv4,
		IPv6:         c.IPv6,
	}, nil
}

// maxRespSize is the maximum size of response from Consul key-value storage.
const maxRespSize = 1 * datasize.MB

// keyNamespaceCheck is the namespace added to the keys of DNS check.  See
// [remotekv.KeyNamespace].
const keyNamespaceCheck = "check"

// newRemoteKV returns a new properly initialized remote key-value storage.  c
// must be valid.  grpcMtrc should be registered before calling this method.
func (c *remoteKVConfig) newRemoteKV(
	envs *environment,
	namespace string,
	reg prometheus.Registerer,
	grpcMtrc backendpb.GRPCMetrics,
	baseLogger *slog.Logger,
) (kv remotekv.Interface, err error) {
	switch c.Type {
	case kvModeBackend:
		return newBackendRemoteKV(envs, namespace, reg, grpcMtrc, c.TTL)
	case kvModeCache:
		// TODO(e.burkov): The local cache in [dnscheck.RemoteKV] becomes
		// pointless with this mode.
		return remotekv.NewCache(&remotekv.CacheConfig{
			Cache: agdcache.NewLRU[string, []byte](&agdcache.LRUConfig{
				Count: envs.DNSCheckCacheKVSize,
			}),
		}), nil
	case kvModeRedis:
		return newRedisRemoteKV(envs, namespace, reg, baseLogger, c.TTL)
	case kvModeConsul:
		kv, err = newConsulRemoteKV(envs, c.TTL)
		if err != nil {
			return nil, fmt.Errorf("initializing consul dnscheck kv: %w", err)
		}
	default:
		panic(fmt.Errorf("dnscheck kv type: %w: %q", errors.ErrBadEnumValue, c.Type))
	}

	return remotekv.NewKeyNamespace(&remotekv.KeyNamespaceConfig{
		KV:     kv,
		Prefix: newRemoteKVPrefix(envs, c.Type),
	}), nil
}

// newBackendRemoteKV returns a new properly initialized backend remote
// key-value storage.
//
// TODO(e.burkov):  Add key namespace.
func newBackendRemoteKV(
	envs *environment,
	namespace string,
	reg prometheus.Registerer,
	grpcMtrc backendpb.GRPCMetrics,
	ttl timeutil.Duration,
) (remotekv.Interface, error) {
	backendKVMtrc, err := metrics.NewBackendRemoteKV(namespace, reg)
	if err != nil {
		return nil, fmt.Errorf("registering backend kv metrics: %w", err)
	}

	var kv *backendpb.RemoteKV
	kv, err = backendpb.NewRemoteKV(&backendpb.RemoteKVConfig{
		GRPCMetrics: grpcMtrc,
		Metrics:     backendKVMtrc,
		Endpoint:    &envs.DNSCheckRemoteKVURL.URL,
		APIKey:      envs.DNSCheckRemoteKVAPIKey,
		TTL:         time.Duration(ttl),
	})
	if err != nil {
		return nil, fmt.Errorf("initializing backend dnscheck kv: %w", err)
	}

	return kv, nil
}

// newRedisRemoteKV returns a new properly initialized Redis-based remote
// key-value storage.
func newRedisRemoteKV(
	envs *environment,
	namespace string,
	reg prometheus.Registerer,
	baseLogger *slog.Logger,
	ttl timeutil.Duration,
) (remotekv.Interface, error) {
	mtrc, err := metrics.NewRedisKV(namespace, reg)
	if err != nil {
		return nil, fmt.Errorf("registering redis kv metrics: %w", err)
	}

	var dialer *redisutil.DefaultDialer
	dialer, err = redisutil.NewDefaultDialer(&redisutil.DefaultDialerConfig{
		Addr: &netutil.HostPort{
			Host: envs.RedisAddr,
			Port: envs.RedisPort,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initializing redisutil dialer: %w", err)
	}

	var connTester *redisutil.RoleChecker
	connTester, err = redisutil.NewRoleChecker(&redisutil.RoleCheckerConfig{
		Logger: baseLogger.With(slogutil.KeyPrefix, "redis_role_checker"),
	})
	if err != nil {
		return nil, fmt.Errorf("initializing redisutil role checker: %w", err)
	}

	var pool *redisutil.DefaultPool
	pool, err = redisutil.NewDefaultPool(&redisutil.DefaultPoolConfig{
		Logger:           baseLogger.With(slogutil.KeyPrefix, "redis_pool"),
		ConnectionTester: connTester,
		Dialer:           dialer,
		Metrics:          mtrc,
		IdleTimeout:      time.Duration(envs.RedisIdleTimeout),
		MaxActive:        envs.RedisMaxActive,
		MaxIdle:          envs.RedisMaxIdle,
		Wait:             true,
	})
	if err != nil {
		return nil, fmt.Errorf("initializing redisutil pool: %w", err)
	}

	kv := rediskv.NewRedisKV(&rediskv.RedisKVConfig{
		Pool: pool,
		TTL:  time.Duration(ttl),
	})

	return kv, nil
}

// newConsulRemoteKV returns a new properly initialized Consul-based remote
// key-value storage.
func newConsulRemoteKV(envs *environment, ttl timeutil.Duration) (kv remotekv.Interface, err error) {
	consulKVURL := envs.ConsulDNSCheckKVURL
	consulSessionURL := envs.ConsulDNSCheckSessionURL
	if consulKVURL == nil || consulSessionURL == nil {
		return remotekv.Empty{}, nil
	}

	kv, err = consulkv.NewKV(&consulkv.Config{
		URL:        &consulKVURL.URL,
		SessionURL: &consulSessionURL.URL,
		Client: agdhttp.NewClient(&agdhttp.ClientConfig{
			// TODO(ameshkov): Consider making configurable.
			Timeout: 15 * time.Second,
		}),
		// TODO(ameshkov): Consider making configurable.
		Limiter:     rate.NewLimiter(rate.Limit(200)/60, 1),
		TTL:         time.Duration(ttl),
		MaxRespSize: maxRespSize,
	})
	if err != nil {
		return nil, fmt.Errorf("initializing consul dnscheck kv: %w", err)
	}

	return kv, nil
}

// newRemoteKVPrefix returns a remote KV custom prefix for the keys.
func newRemoteKVPrefix(envs *environment, kvType string) (pref string) {
	switch kvType {
	case kvModeBackend, kvModeCache, kvModeConsul:
		return fmt.Sprintf("%s:%s:", kvType, keyNamespaceCheck)
	case kvModeRedis:
		return fmt.Sprintf("%s:%s:", envs.RedisKeyPrefix, keyNamespaceCheck)
	default:
		panic(fmt.Errorf("dnscheck kv type: %w: %q", errors.ErrBadEnumValue, kvType))
	}
}

// type check
var _ validate.Interface = (*checkConfig)(nil)

// Validate implements the [validate.Interface] interface for *checkConfig.
func (c *checkConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	errs := []error{
		validate.NotEmpty("node_location", c.NodeLocation),
		validate.NotEmpty("node_name", c.NodeName),
		validate.NotEmptySlice("domains", c.Domains),
	}

	err = validateNonNilIPs(c.IPv4, netutil.AddrFamilyIPv4)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		errs = append(errs, err)
	}

	err = validateNonNilIPs(c.IPv6, netutil.AddrFamilyIPv6)
	if err != nil {
		// Don't wrap the error, because it's informative enough as is.
		errs = append(errs, err)
	}

	errs = validate.Append(errs, "kv", c.KV)

	return errors.Join(errs...)
}

// ValidateNonNilIPs returns an error if ips is empty or had IP addresses of
// incorrect protocol version.
//
// TODO(a.garipov): Merge with [validateAddrs].
func validateNonNilIPs(ips []netip.Addr, fam netutil.AddrFamily) (err error) {
	if len(ips) == 0 {
		return fmt.Errorf("no %s", fam)
	}

	// Assume that since ips are parsed from YAML, they are valid.

	var checkProto func(ip netip.Addr) (ok bool)
	switch fam {
	case netutil.AddrFamilyIPv4:
		checkProto = netip.Addr.Is4
	case netutil.AddrFamilyIPv6:
		checkProto = netip.Addr.Is6
	default:
		panic(fmt.Errorf("agdnet: unsupported addr fam %s", fam))
	}

	for i, ip := range ips {
		if !checkProto(ip) {
			return fmt.Errorf("%s: address %q at index %d: incorrect protocol", fam, ip, i)
		}
	}

	return nil
}

// DNSCheck key-value database modes.
const (
	kvModeBackend = "backend"
	kvModeCache   = "cache"
	kvModeConsul  = "consul"
	kvModeRedis   = "redis"
)

// remoteKVConfig is remote key-value store configuration for DNS server
// checking.
type remoteKVConfig struct {
	// Type defines the type of remote key-value store.  Allowed values are
	// [kvModeBackend], [kvModeCache], [kvModeConsul] and [kvModeRedis].
	Type string `yaml:"type"`

	// TTL defines, for how long to keep the information about a single client.
	TTL timeutil.Duration `yaml:"ttl"`
}

// type check
var _ validate.Interface = (*remoteKVConfig)(nil)

// Validate implements the [validate.Interface] interface for *remoteKVConfig.
func (c *remoteKVConfig) Validate() (err error) {
	if c == nil {
		return errors.ErrNoValue
	}

	ttl := time.Duration(c.TTL)

	switch c.Type {
	case kvModeBackend:
		return validate.Positive("ttl", ttl)
	case kvModeCache:
		return nil
	case kvModeConsul:
		return validate.InRange("ttl", ttl, consulkv.MinTTL, consulkv.MaxTTL)
	case kvModeRedis:
		return validate.NoLessThan("ttl", ttl, rediskv.MinTTL)
	default:
		return fmt.Errorf("type: %w: %q", errors.ErrBadEnumValue, c.Type)
	}
}
