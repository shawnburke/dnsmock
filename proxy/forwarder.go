package proxy

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/shawnburke/dnsmock/config"
	"go.uber.org/zap"
)

type Forwarder interface {
	Forward(msg *dns.Msg) (*dns.Msg, error)
}

func NewForwarder(cfg config.Parameters, logger *zap.Logger) Forwarder {

	fi := &forwarderImpl{
		logger: logger,
		searchNames: func(name string) []string {
			return []string{name}
		},
		local: cfg.IsDownstreamLocalhost(),
	}

	dnsClient := &dns.Client{
		Net:          "udp",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	fi.client = dnsClient
	fi.servers = normalizeServers(cfg.Downstreams()...)
	return fi
}

type forwarderImpl struct {
	sync.RWMutex
	servers              []string
	logger               *zap.Logger
	client               *dns.Client
	local                bool
	resolveConfTimestamp time.Time
	searchNames          func(string) []string
}

func normalizeServers(servers ...string) []string {

	for i, s := range servers {

		c := strings.Index(s, ":")
		if c == -1 {
			s += ":53"
		}
		servers[i] = s
	}
	return servers
}

const resolveConfPath = "/etc/resolv.conf"

func (f *forwarderImpl) getServers() []string {
	servers := f.servers

	if f.local {

		rc, err := os.Stat(resolveConfPath)

		if err != nil {
			f.logger.Debug("Can't stat /etc/resolv.conf", zap.Error(err))
			return servers
		}

		lastMod := rc.ModTime()

		f.Lock()
		defer f.Unlock()

		if lastMod != f.resolveConfTimestamp {
			f.resolveConfTimestamp = lastMod
			local, err := dns.ClientConfigFromFile(resolveConfPath)
			if err != nil {
				f.logger.Panic("Can't load /etc/resolv.conf", zap.Error(err))
			}
			f.servers = normalizeServers(local.Servers...)
			f.searchNames = local.NameList
		}
		servers = f.servers
	}

	return servers
}

func (f *forwarderImpl) Forward(msg *dns.Msg) (*dns.Msg, error) {
	servers := f.getServers()

	name := msg.Question[0].Name
	names := f.searchNames(dns.CanonicalName(name))
	response := &dns.Msg{}
Outer:
	for _, server := range servers {
		for _, n := range names {
			msg.Question[0].Name = n
			res, err := f.forwardCore(server, msg)
			if err != nil {
				continue
			}
			if len(res.Answer) > 0 {
				response = res
				break Outer
			}
		}
	}
	msg.Question[0].Name = name
	response.SetReply(msg)
	return response, nil
}

func (f *forwarderImpl) forwardCore(server string, m *dns.Msg) (*dns.Msg, error) {
	f.logger.Debug("Forwarding DNS request", zap.String("server", server), zap.String("question", m.Question[0].String()))
	response, _, err := f.client.Exchange(m, server)
	if err != nil {
		f.logger.Error("Failed to forward DNS request", zap.Error(err), zap.String("server", server), zap.String("question", m.Question[0].String()))
		return nil, err
	}
	f.logger.Debug("Forwarded DNS request", zap.String("server", server), zap.String("question", m.Question[0].String()), zap.String("response", response.String()))
	return response, nil
}
