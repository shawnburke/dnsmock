package resolver

import (
	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/spec"
	"go.uber.org/zap"
)

// NewReplay creates a resolver that replays canned messages
func NewReplay(r *spec.Responses, logger *zap.Logger) Resolver {
	return &replayResolver{responses: r, logger: logger.With(zap.String("resolver", "replay"))}
}

func NewReplayFromFile(p string, logger *zap.Logger) Resolver {
	s := spec.FromFile(p)

	return NewReplay(s, logger)
}

type replayResolver struct {
	responses *spec.Responses
	logger    *zap.Logger
}

func (r *replayResolver) Resolve(msg *dns.Msg) (*dns.Msg, error) {
	response := r.responses.Find(msg)
	if response != nil {
		r.logger.Debug(
			"REPLAY-RESOLVER: replaying response",
			zap.String("question", msg.Question[0].String()),
			zap.Strings("answer", AnswerStrings(response)),
		)

	} else {
		r.logger.Debug("REPLAY-RESOLVER: no response found", zap.String("question", msg.Question[0].String()))
	}
	return response, nil
}
