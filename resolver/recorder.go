package resolver

import (
	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/spec"
	"go.uber.org/zap"
)

// NewRecorder creates a resolver that records responses from another resolver
func NewRecorder(r Resolver, responses *spec.Responses, logger *zap.Logger) Resolver {
	return &recorderResolver{
		resolver:  r,
		responses: responses,
		logger:    logger.With(zap.String("resolver", "recorder")),
	}
}

type recorderResolver struct {
	resolver  Resolver
	responses *spec.Responses
	logger    *zap.Logger
}

func (r *recorderResolver) Resolve(msg *dns.Msg) (*dns.Msg, error) {
	response, err := r.resolver.Resolve(msg)
	if err != nil {
		return nil, err
	}
	if response != nil && len(response.Answer) > 0 {
		r.logger.Debug("RECORDER-RESOLVER: recording response",
			zap.String("question", msg.Question[0].String()),
			zap.Strings("answer", AnswerStrings(response)),
		)
		r.responses.Add(msg, response)
	}
	return response, nil
}
