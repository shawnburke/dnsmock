package dnsmock

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/resolver"
	"github.com/shawnburke/dnsmock/spec"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// This is an example E22 usage of the library where
// we create a DNS server and mock out Google DNS
// to 4.3.2.1

const specYaml = `
rules:
 - name: google.com.
   records:
    A:
    - "google.com. 300 IN A 4.3.2.1"
`

func TestMock(t *testing.T) {

	logger := zap.NewNop()

	// spec is our wrapper around the YAML config
	s := spec.FromYAML(specYaml)

	// a replay resolver resolves DNS using the spec
	resolver := resolver.NewReplay(s, logger)

	// finally, createa  proxy that uses the spec resolver
	// only
	proxy := New(":0", resolver, logger)

	err := proxy.Start()
	require.NoError(t, err)
	defer proxy.Stop()

	// test it using the miekg/dns library
	client := &dns.Client{
		Net: "udp",
	}

	query := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "google.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
	}

	res, _, err := client.Exchange(query, proxy.Addr())

	require.NoError(t, err)

	require.Equal(t, 1, len(res.Answer))
	a := res.Answer[0].(*dns.A)
	require.Equal(t, "4.3.2.1", a.A.String())

}
