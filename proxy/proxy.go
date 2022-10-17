package proxy

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New(logger *zap.Logger, cfg config.Parameters, lifecycle fx.Lifecycle) *proxy {

	p := &proxy{
		logger: logger,
		config: cfg,
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

type proxy struct {
	sync.Mutex
	logger *zap.Logger
	config config.Parameters
	server *dns.Server
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

	response, err := p.forward(p.config.Downstreams[0], question)
	if err != nil {
		p.logger.Error("Failed to forward DNS request", zap.Error(err))
		msg := &dns.Msg{}
		msg.SetRcode(question, dns.RcodeServerFailure)
		w.WriteMsg(msg)
		return
	}
	response.SetReply(question)
	w.WriteMsg(response)
}

func (p *proxy) forward(server string, m *dns.Msg) (*dns.Msg, error) {
	dnsClient := new(dns.Client)
	dnsClient.Net = "udp"
	p.logger.Debug("Forwarding DNS request", zap.String("server", server), zap.String("question", m.Question[0].String()))
	response, _, err := dnsClient.Exchange(m, server)
	if err != nil {
		p.logger.Error("Failed to forward DNS request", zap.Error(err), zap.String("server", server), zap.String("question", m.Question[0].String()))
		return nil, err
	}
	p.logger.Debug("Forwarded DNS request", zap.String("server", server), zap.String("question", m.Question[0].String()), zap.String("response", response.String()))
	return response, nil
}

func (p *proxy) Stop() error {
	p.logger.Info("Stopping DNS server")
	return p.server.Shutdown()
}
