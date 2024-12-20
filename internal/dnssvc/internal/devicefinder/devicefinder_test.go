package devicefinder_test

import (
	"context"
	"net/netip"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/AdguardTeam/AdGuardDNS/internal/agd"
	"github.com/AdguardTeam/AdGuardDNS/internal/agdnet"
	"github.com/AdguardTeam/AdGuardDNS/internal/agdtest"
	"github.com/AdguardTeam/AdGuardDNS/internal/dnsmsg"
	"github.com/AdguardTeam/AdGuardDNS/internal/dnsserver"
	"github.com/AdguardTeam/AdGuardDNS/internal/dnsserver/dnsservertest"
	"github.com/AdguardTeam/AdGuardDNS/internal/dnssvc/internal/devicefinder"
	"github.com/AdguardTeam/AdGuardDNS/internal/dnssvc/internal/dnssvctest"
	"github.com/AdguardTeam/AdGuardDNS/internal/profiledb"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	srvPlainWithBindData.SetBindData([]*agd.ServerBindData{{
		ListenConfig: &agdtest.ListenConfig{},
		PrefixAddr: &agdnet.PrefixNetAddr{
			// TODO(a.garipov): Move to dnssvctest?
			Prefix: netip.MustParsePrefix("192.0.2.0/30"),
			Net:    "udp",
			Port:   53,
		},
	}})

	os.Exit(m.Run())
}

// Common requests for tests.
var (
	reqNormal = dnsservertest.NewReq(dnssvctest.DomainFQDN, dns.TypeA, dns.ClassINET)
	reqEDNS   = dnsservertest.NewReq(
		dnssvctest.DomainFQDN,
		dns.TypeA,
		dns.ClassINET,
		dnsservertest.SectionExtra{
			newExtraOPT(1234, []byte{5, 6, 7, 8}),
		},
	)
	reqEDNSDevID = dnsservertest.NewReq(
		dnssvctest.DomainFQDN,
		dns.TypeA,
		dns.ClassINET,
		dnsservertest.SectionExtra{
			newExtraOPT(devicefinder.DnsmasqCPEIDOption, []byte(dnssvctest.DeviceID)),
		},
	)
	reqEDNSBadDevID = dnsservertest.NewReq(
		dnssvctest.DomainFQDN,
		dns.TypeA,
		dns.ClassINET,
		dnsservertest.SectionExtra{
			newExtraOPT(devicefinder.DnsmasqCPEIDOption, []byte("!!!")),
		},
	)
)

// testPassword is the common password for tests.
//
// TODO(a.garipov): Move to dnssvctest?
const testPassword = "123456"

// newExtraOPT returns a new dns.OPT with a local option with the given code and
// data.
func newExtraOPT(code uint16, data []byte) (opt *dns.OPT) {
	return &dns.OPT{
		Hdr: dns.RR_Header{
			Rrtype: dns.TypeOPT,
		},
		Option: []dns.EDNS0{&dns.EDNS0_LOCAL{
			Code: code,
			Data: data,
		}},
	}
}

// Common servers for tests.
var (
	srvPlain = &agd.Server{
		Protocol:        agd.ProtoDNS,
		LinkedIPEnabled: false,
	}
	srvPlainWithLinkedIP = &agd.Server{
		Protocol:        agd.ProtoDNS,
		LinkedIPEnabled: true,
	}
	srvDoH = &agd.Server{
		Protocol: agd.ProtoDoH,
	}
	srvDoQ = &agd.Server{
		Protocol: agd.ProtoDoQ,
	}
	srvDoT = &agd.Server{
		Protocol: agd.ProtoDoT,
	}

	// NOTE:  The bind data are set in [TestMain].
	srvPlainWithBindData = &agd.Server{
		Protocol:        agd.ProtoDNS,
		LinkedIPEnabled: false,
	}
)

// Common profiles, devices, and results for tests.
var (
	profNormal = &agd.Profile{
		BlockingMode: &dnsmsg.BlockingModeNullIP{},
		ID:           dnssvctest.ProfileID,
		DeviceIDs:    []agd.DeviceID{dnssvctest.DeviceID},
		Deleted:      false,
	}

	profDeleted = &agd.Profile{
		BlockingMode: &dnsmsg.BlockingModeNullIP{},
		ID:           dnssvctest.ProfileID,
		DeviceIDs:    []agd.DeviceID{dnssvctest.DeviceID},
		Deleted:      true,
	}

	devNormal = &agd.Device{
		Auth: &agd.AuthSettings{
			Enabled: false,
		},
		ID:       dnssvctest.DeviceID,
		LinkedIP: dnssvctest.LinkedAddr,
	}

	devAuto = &agd.Device{
		Auth: &agd.AuthSettings{
			Enabled: false,
		},
		ID:           dnssvctest.DeviceID,
		HumanIDLower: dnssvctest.HumanIDLower,
	}

	resNormal = &agd.DeviceResultOK{
		Device:  devNormal,
		Profile: profNormal,
	}
)

// newDevAuth returns a new device with the given parameters for tests.
func newDevAuth(dohAuthOnly, passwdMatches bool) (d *agd.Device) {
	return &agd.Device{
		Auth: &agd.AuthSettings{
			PasswordHash: &agdtest.Authenticator{
				OnAuthenticate: func(_ context.Context, _ []byte) (ok bool) {
					return passwdMatches
				},
			},
			Enabled:     true,
			DoHAuthOnly: dohAuthOnly,
		},
		ID: dnssvctest.DeviceID,
	}
}

// newOnProfileByDedicatedIP returns a function with the type of
// [agdtest.ProfileDB.OnProfileByDedicatedIP] that returns p and d only when
// localIP is equal to the given one.
func newOnProfileByDedicatedIP(
	wantLocalIP netip.Addr,
) (f func(_ context.Context, localIP netip.Addr) (p *agd.Profile, d *agd.Device, err error)) {
	return func(_ context.Context, localIP netip.Addr) (p *agd.Profile, d *agd.Device, err error) {
		if localIP == wantLocalIP {
			return profNormal, devNormal, nil
		}

		return nil, nil, profiledb.ErrDeviceNotFound
	}
}

// newOnProfileByDeviceID returns a function with the type of
// [agdtest.ProfileDB.OnProfileByDeviceID] that returns p and d only when devID
// is equal to the given one.
func newOnProfileByDeviceID(
	wantDevID agd.DeviceID,
) (f func(_ context.Context, devID agd.DeviceID) (p *agd.Profile, d *agd.Device, err error)) {
	return func(_ context.Context, devID agd.DeviceID) (p *agd.Profile, d *agd.Device, err error) {
		if devID == wantDevID {
			return profNormal, devNormal, nil
		}

		return nil, nil, profiledb.ErrDeviceNotFound
	}
}

// newOnProfileByHumanID returns a function with the type of
// [agdtest.ProfileDB.OnProfileByHumanID] that returns p and d only when id and
// humanID are equal to the given one.
func newOnProfileByHumanID(
	wantProfID agd.ProfileID,
	wantHumanID agd.HumanIDLower,
) (
	f func(
		_ context.Context,
		id agd.ProfileID,
		humanID agd.HumanIDLower,
	) (p *agd.Profile, d *agd.Device, err error),
) {
	return func(
		_ context.Context,
		id agd.ProfileID,
		humanID agd.HumanIDLower,
	) (p *agd.Profile, d *agd.Device, err error) {
		if id == wantProfID && humanID == wantHumanID {
			return profNormal, devAuto, nil
		}

		return nil, nil, profiledb.ErrDeviceNotFound
	}
}

// newOnProfileByLinkedIP returns a function with the type of
// [agdtest.ProfileDB.OnProfileByLinkedIP] that returns p and d only when
// remoteIP is equal to the given one.
func newOnProfileByLinkedIP(
	wantRemoteIP netip.Addr,
) (f func(_ context.Context, remoteIP netip.Addr) (p *agd.Profile, d *agd.Device, err error)) {
	return func(_ context.Context, remoteIP netip.Addr) (p *agd.Profile, d *agd.Device, err error) {
		if remoteIP == wantRemoteIP {
			return profNormal, devNormal, nil
		}

		return nil, nil, profiledb.ErrDeviceNotFound
	}
}

// assertEqualResult is a helper that uses [assert.Equal] for all result types
// except [*agd.DeviceResultError], for which it uses [testutil.AssertErrorMsg].
func assertEqualResult(tb testing.TB, want, got agd.DeviceResult) {
	tb.Helper()

	switch want := want.(type) {
	case *agd.DeviceResultError:
		gotRE := testutil.RequireTypeAssert[*agd.DeviceResultError](tb, got)
		testutil.AssertErrorMsg(tb, want.Err.Error(), gotRE.Err)
	default:
		assert.Equal(tb, want, got)
	}
}

func TestDefault_Find_dnscrypt(t *testing.T) {
	t.Parallel()

	df := devicefinder.NewDefault(&devicefinder.Config{
		Logger:        slogutil.NewDiscardLogger(),
		HumanIDParser: agd.NewHumanIDParser(),
		Server: &agd.Server{
			Protocol: agd.ProtoDNSCrypt,
		},
	})

	ctx := testutil.ContextWithTimeout(t, dnssvctest.Timeout)
	r := df.Find(ctx, reqNormal, dnssvctest.ClientAddrPort, dnssvctest.ServerAddrPort)
	assert.Nil(t, r)
}

// Common sinks for benchmarks.
var (
	sinkDevResult agd.DeviceResult
)

func BenchmarkDefault(b *testing.B) {
	profDB := &agdtest.ProfileDB{
		OnCreateAutoDevice: func(
			_ context.Context,
			_ agd.ProfileID,
			_ agd.HumanID,
			_ agd.DeviceType,
		) (p *agd.Profile, d *agd.Device, err error) {
			panic("not implemented")
		},

		OnProfileByDedicatedIP: func(
			_ context.Context,
			_ netip.Addr,
		) (p *agd.Profile, d *agd.Device, err error) {
			return profNormal, devNormal, nil
		},

		OnProfileByDeviceID: func(
			_ context.Context,
			_ agd.DeviceID,
		) (p *agd.Profile, d *agd.Device, err error) {
			return profNormal, devNormal, nil
		},

		OnProfileByHumanID: func(
			_ context.Context,
			_ agd.ProfileID,
			_ agd.HumanIDLower,
		) (p *agd.Profile, d *agd.Device, err error) {
			return profNormal, devNormal, nil
		},

		OnProfileByLinkedIP: func(
			_ context.Context,
			_ netip.Addr,
		) (p *agd.Profile, d *agd.Device, err error) {
			return profNormal, devNormal, nil
		},
	}

	benchCases := []struct {
		conf       *devicefinder.Config
		req        *dns.Msg
		srvReqInfo *dnsserver.RequestInfo
		name       string
	}{{
		conf: &devicefinder.Config{
			Logger:        slogutil.NewDiscardLogger(),
			ProfileDB:     profDB,
			HumanIDParser: agd.NewHumanIDParser(),
			Server:        srvDoT,
			DeviceDomains: []string{dnssvctest.DomainForDevices},
		},
		req: reqNormal,
		srvReqInfo: &dnsserver.RequestInfo{
			TLSServerName: dnssvctest.DeviceIDSrvName,
		},
		name: "dot",
	}, {
		conf: &devicefinder.Config{
			Logger:        slogutil.NewDiscardLogger(),
			ProfileDB:     profDB,
			HumanIDParser: agd.NewHumanIDParser(),
			Server:        srvDoH,
			DeviceDomains: []string{dnssvctest.DomainForDevices},
		},
		req: reqNormal,
		srvReqInfo: &dnsserver.RequestInfo{
			TLSServerName: dnssvctest.DeviceIDSrvName,
			URL: &url.URL{
				Path: dnsserver.PathDoH,
			},
		},
		name: "doh_domain",
	}, {
		conf: &devicefinder.Config{
			Logger:        slogutil.NewDiscardLogger(),
			ProfileDB:     profDB,
			HumanIDParser: agd.NewHumanIDParser(),
			Server:        srvDoH,
			DeviceDomains: []string{dnssvctest.DomainForDevices},
		},
		req: reqNormal,
		srvReqInfo: &dnsserver.RequestInfo{
			TLSServerName: dnssvctest.DomainForDevices,
			URL: &url.URL{
				Path: path.Join(dnsserver.PathDoH, dnssvctest.DeviceIDStr),
			},
		},
		name: "doh_path",
	}, {
		conf: &devicefinder.Config{
			Logger:        slogutil.NewDiscardLogger(),
			ProfileDB:     profDB,
			HumanIDParser: agd.NewHumanIDParser(),
			Server:        srvPlain,
			DeviceDomains: nil,
		},
		req:        reqEDNSDevID,
		srvReqInfo: &dnsserver.RequestInfo{},
		name:       "dns_edns",
	}, {
		conf: &devicefinder.Config{
			Logger:        slogutil.NewDiscardLogger(),
			ProfileDB:     profDB,
			HumanIDParser: agd.NewHumanIDParser(),
			Server:        srvPlainWithBindData,
			DeviceDomains: nil,
		},
		req:        reqNormal,
		srvReqInfo: &dnsserver.RequestInfo{},
		name:       "dns_laddr",
	}, {
		conf: &devicefinder.Config{
			Logger:        slogutil.NewDiscardLogger(),
			ProfileDB:     profDB,
			HumanIDParser: agd.NewHumanIDParser(),
			Server:        srvPlainWithLinkedIP,
			DeviceDomains: nil,
		},
		req:        reqNormal,
		srvReqInfo: &dnsserver.RequestInfo{},
		name:       "dns_raddr",
	}}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			df := devicefinder.NewDefault(bc.conf)

			ctx := testutil.ContextWithTimeout(b, dnssvctest.Timeout)
			ctx = dnsserver.ContextWithRequestInfo(ctx, bc.srvReqInfo)

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				sinkDevResult = df.Find(
					ctx,
					bc.req,
					dnssvctest.ClientAddrPort,
					dnssvctest.ServerAddrPort,
				)
			}

			_ = testutil.RequireTypeAssert[*agd.DeviceResultOK](b, sinkDevResult)
		})
	}

	// Most recent result:
	//	goos: linux
	//	goarch: amd64
	//	pkg: github.com/AdguardTeam/AdGuardDNS/internal/dnssvc/internal/devicefinder
	//	cpu: AMD Ryzen 7 PRO 4750U with Radeon Graphics
	//	BenchmarkDefault/dot-16         	 5258900	       300.3 ns/op	      16 B/op	       1 allocs/op
	//	BenchmarkDefault/doh_domain-16  	 1996458	       621.0 ns/op	      64 B/op	       3 allocs/op
	//	BenchmarkDefault/doh_path-16    	 2376877	       655.0 ns/op	      80 B/op	       3 allocs/op
	//	BenchmarkDefault/dns_edns-16    	 4566312	       289.3 ns/op	      24 B/op	       2 allocs/op
	//	BenchmarkDefault/dns_laddr-16   	 6154356	       198.7 ns/op	      16 B/op	       1 allocs/op
	//	BenchmarkDefault/dns_raddr-16   	 7268647	       183.3 ns/op	      16 B/op	       1 allocs/op
}
