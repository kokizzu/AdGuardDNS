package safesearch_test

import (
	"net"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardDNS/internal/agdtest"
	"github.com/AdguardTeam/AdGuardDNS/internal/dnsmsg"
	"github.com/AdguardTeam/AdGuardDNS/internal/dnsserver/dnsservertest"
	"github.com/AdguardTeam/AdGuardDNS/internal/filter"
	"github.com/AdguardTeam/AdGuardDNS/internal/filter/internal/filtertest"
	"github.com/AdguardTeam/AdGuardDNS/internal/filter/internal/refreshable"
	"github.com/AdguardTeam/AdGuardDNS/internal/filter/internal/rulelist"
	"github.com/AdguardTeam/AdGuardDNS/internal/filter/internal/safesearch"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSafeIPStr is the string representation of the IP address of the safe
// version of [testEngineWithIP].
const testSafeIPStr = "1.2.3.4"

// testIPOfEngineWithIP is the IP address of the safe version of
// search-engine-ip.example.
var testIPOfEngineWithIP = netip.MustParseAddr(testSafeIPStr)

// Common domain names for tests.
const (
	testOther            = "other.example"
	testEngineWithIP     = "search-engine-ip.example"
	testEngineWithDomain = "search-engine-domain.example"
	testSafeDomain       = "safe-search-engine-domain.example"
)

// testFilterRules is are common filtering rules for tests.
const testFilterRules = `|` + testEngineWithIP + `^$dnsrewrite=NOERROR;A;` + testSafeIPStr + "\n" +
	`|` + testEngineWithDomain + `^$dnsrewrite=NOERROR;CNAME;` + testSafeDomain

func TestFilter(t *testing.T) {
	reqCh := make(chan struct{}, 1)
	cachePath, srvURL := filtertest.PrepareRefreshable(t, reqCh, testFilterRules, http.StatusOK)

	f, newErr := safesearch.New(
		&safesearch.Config{
			Refreshable: &refreshable.Config{
				Logger:    slogutil.NewDiscardLogger(),
				ID:        filter.IDGeneralSafeSearch,
				URL:       srvURL,
				CachePath: cachePath,
				Staleness: filtertest.Staleness,
				Timeout:   filtertest.Timeout,
				MaxSize:   filtertest.FilterMaxSize,
			},
			CacheTTL: 1 * time.Minute,
		},
		rulelist.NewResultCache(filtertest.CacheCount, true),
	)
	require.NoError(t, newErr)

	refrErr := f.Refresh(testutil.ContextWithTimeout(t, filtertest.Timeout), true)
	require.NoError(t, refrErr)

	testutil.RequireReceive(t, reqCh, filtertest.Timeout)

	require.True(t, t.Run("no_match", func(t *testing.T) {
		ctx := testutil.ContextWithTimeout(t, filtertest.Timeout)
		req := newReq(t, testOther, dns.TypeA)
		res, err := f.FilterRequest(ctx, req)
		require.NoError(t, err)

		assert.Nil(t, res)

		require.True(t, t.Run("cached", func(t *testing.T) {
			res, err = f.FilterRequest(ctx, req)
			require.NoError(t, err)

			// TODO(a.garipov): Find a way to make caches more inspectable.
			assert.Nil(t, res)
		}))
	}))

	require.True(t, t.Run("txt", func(t *testing.T) {
		ctx := testutil.ContextWithTimeout(t, filtertest.Timeout)
		req := newReq(t, testEngineWithIP, dns.TypeTXT)
		res, err := f.FilterRequest(ctx, req)
		require.NoError(t, err)

		assert.Nil(t, res)
	}))

	require.True(t, t.Run("ip", func(t *testing.T) {
		ctx := testutil.ContextWithTimeout(t, filtertest.Timeout)
		req := newReq(t, testEngineWithIP, dns.TypeA)
		res, err := f.FilterRequest(ctx, req)
		require.NoError(t, err)

		rm := testutil.RequireTypeAssert[*filter.ResultModifiedResponse](t, res)
		require.Len(t, rm.Msg.Answer, 1)

		assert.Equal(t, rm.Rule, filter.RuleText(testEngineWithIP))

		a := testutil.RequireTypeAssert[*dns.A](t, rm.Msg.Answer[0])
		assert.Equal(t, net.IP(testIPOfEngineWithIP.AsSlice()), a.A)

		t.Run("cached", func(t *testing.T) {
			newReq := newReq(t, testEngineWithIP, dns.TypeA)

			var cachedRes filter.Result
			cachedRes, err = f.FilterRequest(ctx, newReq)
			require.NoError(t, err)

			// Do not assert that the results are the same, since a modified
			// result of a safe search is always cloned.  But assert that the
			// non-clonable fields are equal and that the message has reply
			// fields set properly.
			cachedMR := testutil.RequireTypeAssert[*filter.ResultModifiedResponse](t, cachedRes)
			assert.NotSame(t, cachedMR, rm)
			assert.Equal(t, cachedMR.Msg.Id, newReq.DNS.Id)
			assert.Equal(t, cachedMR.List, rm.List)
			assert.Equal(t, cachedMR.Rule, rm.Rule)
		})
	}))

	require.True(t, t.Run("domain", func(t *testing.T) {
		ctx := testutil.ContextWithTimeout(t, filtertest.Timeout)
		req := newReq(t, testEngineWithDomain, dns.TypeA)
		res, err := f.FilterRequest(ctx, req)
		require.NoError(t, err)

		rm := testutil.RequireTypeAssert[*filter.ResultModifiedRequest](t, res)
		require.NotNil(t, rm.Msg)
		require.Len(t, rm.Msg.Question, 1)

		assert.False(t, rm.Msg.Response)
		assert.Equal(t, rm.Rule, filter.RuleText(testEngineWithDomain))

		q := rm.Msg.Question[0]
		assert.Equal(t, dns.TypeA, q.Qtype)
		assert.Equal(t, dns.Fqdn(testSafeDomain), q.Name)
	}))

	require.True(t, t.Run("https", func(t *testing.T) {
		ctx := testutil.ContextWithTimeout(t, filtertest.Timeout)
		req := newReq(t, testEngineWithDomain, dns.TypeHTTPS)
		res, err := f.FilterRequest(ctx, req)
		require.NoError(t, err)

		rm := testutil.RequireTypeAssert[*filter.ResultModifiedRequest](t, res)
		require.NotNil(t, rm.Msg)
		require.Len(t, rm.Msg.Question, 1)

		assert.False(t, rm.Msg.Response)
		assert.Equal(t, rm.Rule, filter.RuleText(testEngineWithDomain))

		q := rm.Msg.Question[0]
		assert.Equal(t, dns.TypeHTTPS, q.Qtype)
		assert.Equal(t, dns.Fqdn(testSafeDomain), q.Name)
	}))
}

// newReq is a test helper that returns the filtering request with the given
// data.
func newReq(tb testing.TB, host string, qt dnsmsg.RRType) (req *filter.Request) {
	tb.Helper()

	return &filter.Request{
		DNS:      dnsservertest.NewReq(host, qt, dns.ClassINET),
		Messages: agdtest.NewConstructor(tb),
		Host:     host,
		QType:    qt,
		QClass:   dns.ClassINET,
	}
}
