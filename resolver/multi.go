package resolver

import "github.com/miekg/dns"

func NewMulti(resolvers ...Resolver) Resolver {
	return &multiResolver{resolvers: resolvers}
}

type multiResolver struct {
	resolvers []Resolver
}

func (r *multiResolver) Resolve(msg *dns.Msg) (*dns.Msg, error) {
	for _, resolver := range r.resolvers {
		response, err := resolver.Resolve(msg)
		if err != nil {
			continue
		}
		if response != nil && len(response.Answer) > 0 {
			return response, nil
		}
	}
	return nil, nil
}
