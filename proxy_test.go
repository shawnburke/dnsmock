package dnsmock

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/resolver"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var logger = zap.NewNop()

func TestProxySimple(t *testing.T) {

	resolver := resolver.NewDns("8.8.8.8", logger)

	p := New("", resolver, logger)

	err := p.Start()
	defer p.Stop()
	require.NoError(t, err)

	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{},
		Question: []dns.Question{
			{
				Name:   "google.com.",
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
