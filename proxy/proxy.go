package proxy

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/config"
	"github.com/shawnburke/dnsmock/spec"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type proxy struct {
	sync.Mutex
	logger    *zap.Logger
	config    config.Parameters
	server    *dns.Server
	forwarder Forwarder
	responses spec.Responses

	useCache bool
}

func New(
	logger *zap.Logger,
	cfg config.Parameters,
	responses spec.Responses,
	forwarder Forwarder,
	lifecycle fx.Lifecycle) *proxy {

	p := &proxy{
		logger:    logger,
		config:    cfg,
		forwarder: forwarder,
		responses: responses,
		useCache:  cfg.Record || responses.Count() > 0,
	}

	if lifecycle != nil {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {

				return p.Start()
			},
			OnStop: func(ctx context.Context) error {
				return p.Stop()
			},
		})
	}

	return p
}

func (p *proxy) Start() error {
	p.Lock()
	defer p.Unlock()
	var err error

	if p.server != nil {
		return errors.New("AlreadyStarted")
	}

	p.server = &dns.Server{Addr: p.config.ListenAddr, Net: "udp"}
	p.server.Handler = dns.HandlerFunc(p.handler)

	go func() {

		p.logger.Info("Starting DNS server", zap.String("addr", p.config.ListenAddr))
		err := p.server.ListenAndServe()
		if err != nil {
			p.logger.Error("Failed to start DNS server", zap.Error(err))
		}
	}()
	time.Sleep(time.Millisecond * 100)
	return err
}

func (p *proxy) handler(w dns.ResponseWriter, question *dns.Msg) {

	var response *dns.Msg

	if p.useCache {
		response = p.responses.Find(question)
	}
	source := "cached"
	if response == nil {
		r2, err := p.forwarder.Forward(question)
		if err != nil {
			p.logger.Error("Failed to forward DNS request", zap.Error(err))
			msg := &dns.Msg{}
			msg.SetRcode(question, dns.RcodeServerFailure)
			w.WriteMsg(msg)
			return
		}
		response = r2
		source = "forwarded"
		if p.config.Record {
			p.responses.Add(question, r2)
		}
	}
	answer := "(none)"
	if len(response.Answer) > 0 {
		answer = response.Answer[0].String()
	}

	p.logger.Debug("Got response",
		zap.String("source", source),
		zap.String("question", question.Question[0].String()),
		zap.Any("first-answer", answer))

	response.SetReply(question)
	w.WriteMsg(response)
}

func (p *proxy) Stop() error {
	p.logger.Info("Stopping DNS server")
	return p.server.Shutdown()
}
