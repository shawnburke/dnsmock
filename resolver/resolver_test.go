package resolver

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/config"
	"github.com/shawnburke/dnsmock/spec"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func makeQuestion(question string, qtype uint16) *dns.Msg {
	m := new(dns.Msg)
	m.SetQuestion(question, qtype)
	return m
}

func getFirstAnswer(res *dns.Msg, qtype uint16) dns.RR {

	for _, a := range res.Answer {
		if a.Header().Rrtype == qtype {
			return a
		}
	}
	return nil
}

func fetch(t *testing.T, q *dns.Msg, r Resolver) dns.RR {
	res, err := r.Resolve(q)
	require.NoError(t, err)
	return getFirstAnswer(res, q.Question[0].Qtype)
}

func TestDns(t *testing.T) {
	r := NewDns("8.8.8.8", zap.NewNop())

	q := makeQuestion("google.com.", dns.TypeA)

	a := fetch(t, q, r)
	require.NotNil(t, a)
}

func TestLocal(t *testing.T) {
	r := NewLocal("", zap.NewNop())
	q := makeQuestion("google.com.", dns.TypeA)
	a := fetch(t, q, r)
	require.NotNil(t, a)
}

var specYaml = `
rules:
 - name: google.com.
   records:
    A:
    - "google.com. 300 IN A 4.3.2.1"
`

func TestReplay(t *testing.T) {

	filename := path.Join(os.TempDir(), fmt.Sprintf("dnsmock-%d.yaml", time.Now().UnixNano()))
	err := os.WriteFile(filename, []byte(specYaml), 0644)
	require.NoError(t, err)
	defer os.Remove(filename)

	s := spec.FromYAML(specYaml)
	r := NewReplay(s, zap.NewNop())

	testReplayCore(t, r)

	r = NewReplayFromFile(filename, zap.NewNop())
	testReplayCore(t, r)
}

func testReplayCore(t *testing.T, r Resolver) {
	q := makeQuestion("google.com.", dns.TypeA)
	a := fetch(t, q, r)
	require.NotNil(t, a)
	ar := a.(*dns.A)
	require.Equal(t, "4.3.2.1", ar.A.String())
}

func TestBuild(t *testing.T) {
	cases := []struct {
		record      bool
		spec        *spec.Responses
		downstreams string
		expected    func(t *testing.T, r Resolver)
	}{
		{
			record:      true,
			spec:        spec.FromYAML(specYaml),
			downstreams: "8.8.8.8,1.1.1.1:53",
			expected: func(t *testing.T, r Resolver) {
				require.IsType(t, &recorderResolver{}, r)
			},
		},
		{
			spec:        spec.FromYAML(specYaml),
			downstreams: "8.8.8.8,1.1.1.1:53",
			expected: func(t *testing.T, r Resolver) {
				multi, ok := r.(*multiResolver)
				require.True(t, ok)
				require.Len(t, multi.resolvers, 3)

				require.IsType(t, &replayResolver{}, multi.resolvers[0])
				require.IsType(t, &dnsResolver{}, multi.resolvers[1])
				require.IsType(t, &dnsResolver{}, multi.resolvers[2])
			},
		},
		{
			spec:        spec.FromYAML(specYaml),
			downstreams: "none",
			expected: func(t *testing.T, r Resolver) {
				multi, ok := r.(*multiResolver)
				require.True(t, ok)
				require.Len(t, multi.resolvers, 1)

				require.IsType(t, &replayResolver{}, multi.resolvers[0])

			},
		},
		{
			spec:        spec.FromYAML(specYaml),
			downstreams: "localhost",
			expected: func(t *testing.T, r Resolver) {
				multi, ok := r.(*multiResolver)
				require.True(t, ok)
				require.Len(t, multi.resolvers, 2)

				require.IsType(t, &replayResolver{}, multi.resolvers[0])
				require.IsType(t, &localResolver{}, multi.resolvers[1])
			},
		},
	}

	for _, c := range cases {
		t.Run("case", func(t *testing.T) {
			cfg := config.Parameters{
				Record:         c.record,
				DownstreamsRaw: c.downstreams,
			}
			r := Build(cfg, c.spec, zap.NewNop())
			require.NotNil(t, r)

			c.expected(t, r)

		})
	}

}

func TestRecorder(t *testing.T) {
	s2 := spec.New()

	rs := spec.FromYAML(specYaml)
	resolver := NewReplay(rs, zap.NewNop())
	r := NewRecorder(resolver, s2, zap.NewNop())

	q := makeQuestion("google.com.", dns.TypeA)
	a := fetch(t, q, r)
	require.NotNil(t, a)
	ar := a.(*dns.A)
	require.Equal(t, "4.3.2.1", ar.A.String())

	require.Len(t, s2.Rules, 1)
	res := s2.Find(q)
	require.NotNil(t, res)
	a = getFirstAnswer(res, q.Question[0].Qtype)
	require.NotNil(t, a)
	require.Equal(t, rs.Rules[0].Records["A"][0],
		strings.Replace(a.String(), "\t", " ", -1))

}

func TestMulti(t *testing.T) {
	r1 := NewLocal("", zap.NewNop())
	r2 := NewLocal("", zap.NewNop())

	r := NewMulti(r1, r2)

	q := makeQuestion("google.com.", dns.TypeA)
	a := fetch(t, q, r)
	require.NotNil(t, a)

}
