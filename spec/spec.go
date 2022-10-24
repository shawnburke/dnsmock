package spec

import (
	"os"
	"strings"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v3"
)

type Responses struct {
	Rules []*Rule `yaml:"rules"`
}
type Rule struct {
	Name    string              `yaml:"name"`
	Records map[string][]string `yaml:"records"`
}

func New() *Responses {
	return &Responses{}
}

func FromFile(path string) *Responses {

	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return FromYAML(string(content))
}

func FromYAML(y string) *Responses {
	r := &Responses{}

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

func (r *Responses) Add(query *dns.Msg, response *dns.Msg) {

	question := query.Question[0]
	domain := question.Name

	qtype := dns.TypeToString[question.Qtype]

	var rule *Rule

	for _, r := range r.Rules {
		if r.Name == domain {
			rule = r
			break
		}
	}

	if rule == nil {
		rule = &Rule{
			Name:    domain,
			Records: map[string][]string{},
		}
		r.Rules = append(r.Rules, rule)
	}

	val := []string{}

	for _, a := range response.Answer {
		val = append(val, a.String())
	}

	rule.Records[qtype] = val
}

func (r Responses) Count() int {
	return len(r.Rules)
}

func normalizeName(d string) string {
	return dns.CanonicalName(d)
}

func (r Responses) FindDomains(domain string) []*Rule {

	domain = normalizeName(domain)

	rules := []*Rule{}

	for _, rule := range r.Rules {
		k := normalizeName(rule.Name)

		if k == domain {
			rules = append(rules, rule)
			continue
		}

		if strings.HasPrefix(k, "*") {
			d := k[1:]

			if strings.HasSuffix(domain, d) {
				rules = append(rules, rule)
				continue
			}
		}
	}

	return rules
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

	d := r.FindDomains(domain)
	if len(d) == 0 {
		return nil
	}

	for _, rule := range r.FindDomains(domain) {
		val := rule.Records[qtype]
		if val != nil {
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
	}

	return nil

}

func (r Responses) YAML() string {
	raw, err := yaml.Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(raw)
}
