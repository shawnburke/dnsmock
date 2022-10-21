package proxy

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestProxy(t *testing.T) {
	logger := zap.NewNop()
	p := New(logger, config.Parameters{
		ListenAddr: "0.0.0.0:0",
	}, nil, nil, nil)
	require.NotNil(t, p)

	err := p.Start()
	require.NoError(t, err)
	p.Stop()

}

func TestPare(t *testing.T) {
	val := "sofi.com.		300	IN	A	172.64.154.149"

	rr, err := dns.NewRR(val)
	require.NoError(t, err)
	require.Equal(t, rr.Header().Rrtype, dns.TypeA)
}
