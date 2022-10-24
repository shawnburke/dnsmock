package resolver

import (
	"strings"
	"time"

	"github.com/miekg/dns"
	"go.uber.org/zap"
)

const timeout = 5 * time.Second

// new DNS resolver
func NewDns(server string, logger *zap.Logger) Resolver {
	r := &dnsResolver{
		server: NormalizeServers(server)[0],
		client: &dns.Client{
			Net:          "udp",
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}
	r.logger = logger.With(zap.String("server", r.server), zap.String("resolver", "dns"))
	return r
}

type dnsResolver struct {
	server string
	client *dns.Client
	logger *zap.Logger
}

func (r *dnsResolver) Resolve(m *dns.Msg) (*dns.Msg, error) {

	response, _, err := r.client.Exchange(m, r.server)
	if err != nil {
		r.logger.Error(
			"DNS-RESOLVER: Failed to forward DNS request",
			zap.Error(err),
			zap.String("question", m.Question[0].String()),
		)
		return nil, err
	}
	r.logger.Debug(
		"DNS-RESOLVER: Forwarded DNS request",
		zap.String("question", m.Question[0].String()),
		zap.String("response", response.String()),
	)
	return response, nil
}

func NormalizeServers(servers ...string) []string {

	for i, s := range servers {

		c := strings.Index(s, ":")
		if c == -1 {
			s += ":53"
		}
		servers[i] = s
	}
	return servers
}
