package resolver

import (
	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/config"
	"github.com/shawnburke/dnsmock/spec"
	"go.uber.org/zap"
)

type Resolver interface {
	Resolve(msg *dns.Msg) (*dns.Msg, error)
}

const DownstreamLocalhost = "localhost"
const DownstreamNone = "none"

func Build(cfg config.Parameters, s *spec.Responses, logger *zap.Logger) Resolver {
	resolvers := []Resolver{}

	// if we are replaying, add the replay resolver
	if s != nil && s.Count() > 0 {
		resolvers = append(resolvers, NewReplay(s, logger))
	}

	// if we have downstreams,
	ds := cfg.Downstreams()

	if len(ds) > 0 {
		for _, d := range ds {
			var r Resolver
			switch d {
			case DownstreamNone:
				continue
			case DownstreamLocalhost:
				r = NewLocal("", logger)
			default:
				resolvers = append(resolvers, NewDns(d, logger))
			}

			if r != nil {
				resolvers = append(resolvers, r)
			}
		}
	}
	all := NewMulti(resolvers...)

	if cfg.Record {
		all = NewRecorder(all, s, logger)
	}
	return all
}

func AnswerStrings(response *dns.Msg) []string {
	answers := []string{}
	for _, a := range response.Answer {
		answers = append(answers, a.String())
	}
	return answers
}
