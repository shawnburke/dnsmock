package dnsmock

import (
	"errors"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/config"
	"github.com/shawnburke/dnsmock/resolver"
	"go.uber.org/zap"
)

type Proxy interface {
	Start() error
	Stop() error
	Addr() string
}

type proxy struct {
	sync.Mutex
	logger   *zap.Logger
	addr     string
	server   *dns.Server
	resolver resolver.Resolver
}

func New(
	addr string,
	resolver resolver.Resolver,
	logger *zap.Logger) Proxy {

	if addr == "" {
		addr = "0.0.0.0:0"
	}

	return &proxy{
		logger:   logger,
		addr:     addr,
		resolver: resolver,
	}

}

func NewFromConfig(cfg config.Parameters, resolver resolver.Resolver, logger *zap.Logger) Proxy {
	return New(cfg.ListenAddr(), resolver, logger)

}

func (p *proxy) Start() error {
	p.Lock()
	defer p.Unlock()
	var err error

	if p.server != nil {
		return errors.New("AlreadyStarted")
	}

	p.server = &dns.Server{Addr: p.addr, Net: "udp"}
	p.server.Handler = dns.HandlerFunc(p.handler)

	go func() {
		err = p.server.ListenAndServe()
		if err != nil {
			p.logger.Error("Failed to start DNS server", zap.Error(err))
		}
	}()
	time.Sleep(time.Millisecond * 100)
	if err == nil {
		p.addr = p.server.PacketConn.LocalAddr().String()
		p.logger.Info("DNS server started", zap.String("addr", p.addr))
	}
	return err
}

func (p *proxy) Addr() string {
	return p.addr
}

func (p *proxy) handler(w dns.ResponseWriter, question *dns.Msg) {

	response, err := p.resolver.Resolve(question)

	msg := &dns.Msg{}
	defer func() {
		response.SetReply(question)
		w.WriteMsg(response)
	}()

	if response == nil {
		response = &dns.Msg{}
	}

	if err != nil {
		p.logger.Error("Failed to handle DNS request", zap.Error(err))
		msg.SetRcode(question, dns.RcodeServerFailure)
		return
	}

	p.logger.Debug("Got response",
		zap.String("question", question.Question[0].String()),
		zap.Any("answer", resolver.AnswerStrings(response)),
	)

}

func (p *proxy) Stop() error {
	p.logger.Info("Stopping DNS server")
	return p.server.Shutdown()
}

func (p *proxy) send(msg *dns.Msg) (*dns.Msg, error) {
	client := &dns.Client{
		Net: "udp",
	}
	res, _, err := client.Exchange(msg, p.server.PacketConn.LocalAddr().String())

	return res, err

}
