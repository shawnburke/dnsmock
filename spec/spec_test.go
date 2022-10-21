package spec

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func findHelper(content string, question dns.Question) []dns.RR {
	r := FromYAML(content)

	msg := &dns.Msg{
		Question: []dns.Question{
			question,
		},
	}
	res := r.Find(msg)
	if res == nil {
		return nil
	}
	return res.Answer
}

var content = `

sofi.com.:
    A:
        - "sofi.com.\t197\tIN\tA\t172.64.154.149"
        - "sofi.com.\t197\tIN\tA\t104.18.33.107"
"*.sofitest.com.":
    A:
        - "{{Name}}\t300\tIN\tA\t1.2.3.4"
"*": 
    SRV:
        - "{{Name}}\t60\tIN\tSRV\t0 100 42 {{Name}}"
`

func TestFromFile(t *testing.T) {

	answer := findHelper(content, dns.Question{
		Name:   "sofi.com.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	})
	require.Len(t, answer, 2)

	a := answer[0].(*dns.A)
	require.Equal(t, "172.64.154.149", a.A.String())

}

func TestWildcard(t *testing.T) {

	answer := findHelper(content, dns.Question{
		Name:   "test.sofitest.com.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	})
	require.Len(t, answer, 1)

	a := answer[0].(*dns.A)
	require.Equal(t, "test.sofitest.com.\t300\tIN\tA\t1.2.3.4", a.String())

}

func TestGlob(t *testing.T) {

	answer := findHelper(content, dns.Question{
		Name:   "tacos.com.",
		Qtype:  dns.TypeSRV,
		Qclass: dns.ClassINET,
	})
	require.Len(t, answer, 1)

	srv := answer[0].(*dns.SRV)
	require.Equal(t, "tacos.com.\t60\tIN\tSRV\t0 100 42 tacos.com.", srv.String())

}
