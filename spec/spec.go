package spec

import (
	"os"
	"strings"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v3"
)

type Responses map[string]map[string][]string

func New() Responses {
	return Responses{}
}

func FromFile(path string) Responses {

	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return FromYAML(string(content))
}

func FromYAML(y string) Responses {
	r := Responses{}

	err := yaml.Unmarshal([]byte(y), &r)
	if err != nil {
		panic(err)
	}
	return r
}

func Parse(val string) Responses {
	r := Responses{}
	err := yaml.Unmarshal([]byte(val), &r)
	if err != nil {
		panic(err)
	}
	return r
}

func (r Responses) Add(query *dns.Msg, response *dns.Msg) {

	question := query.Question[0]
	domain := question.Name

	qtype := dns.TypeToString[question.Qtype]
	if r[domain] == nil {
		r[domain] = map[string][]string{}
	}

	val := []string{}

	for _, a := range response.Answer {
		val = append(val, a.String())
	}

	r[domain][qtype] = val
}

func (r Responses) Count() int {
	return len(r)
}

func normalizeName(d string) string {
	return dns.CanonicalName(d)
}

func (r Responses) FindDomain(domain string) map[string][]string {

	domain = normalizeName(domain)

	for k, v := range r {
		k = normalizeName(k)
		if strings.HasPrefix(k, "*") {
			d := k[1:]

			if strings.HasSuffix(domain, d) {
				return v
			}
		}
	}

	return r[domain]
}

func (r Responses) expand(query dns.Question, val string) string {
	expanded := val

	// TODO: replace with text.template if any more complicated than this
	if strings.Contains(val, "{{Name}}") {
		expanded = strings.ReplaceAll(val, "{{Name}}", query.Name)
	}

	return expanded
}

func (r Responses) Find(query *dns.Msg) *dns.Msg {
	question := query.Question[0]
	domain := question.Name

	qtype := dns.TypeToString[question.Qtype]

	d := r.FindDomain(domain)
	if d == nil {
		return nil
	}

	val := d[qtype]
	if val == nil {
		return nil
	}

	response := &dns.Msg{}
	response.SetReply(query)

	for _, v := range val {

		v2 := r.expand(question, v)

		rr, err := dns.NewRR(v2)
		if err != nil {
			panic(err)
		}
		response.Answer = append(response.Answer, rr)
	}

	return response
}

func (r Responses) YAML() string {
	raw, err := yaml.Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(raw)
}
