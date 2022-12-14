package dnsmock

import (
	"os"
	"testing"

	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/resolver"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var logger = zap.NewNop()

func TestProxySimple(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test on CI")
	}

	resolver := resolver.NewLocal("", logger)

	p := New("", resolver, logger)

	err := p.Start()
	defer p.Stop()
	require.NoError(t, err)

	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{},
		Question: []dns.Question{
			{
				Name:   "github.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
	}

	proxy := p.(*proxy)
	res, err := proxy.send(msg)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.GreaterOrEqual(t, len(res.Answer), 1)
	a := res.Answer[0].(*dns.A)

	require.NotEmpty(t, a.A)

}

func TestParse(t *testing.T) {
	val := "internet.com.		300	IN	A	172.64.154.149"

	rr, err := dns.NewRR(val)
	require.NoError(t, err)
	require.Equal(t, rr.Header().Rrtype, dns.TypeA)
}
