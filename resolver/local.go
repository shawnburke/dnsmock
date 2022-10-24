package resolver

import (
	"os"
	"sync"
	"time"

	"github.com/miekg/dns"
	"go.uber.org/zap"
)

const DefaultResolvConfPath = "/etc/resolv.conf"

func NewLocal(path string, logger *zap.Logger) Resolver {
	if path == "" {
		path = DefaultResolvConfPath
	}
	return &localResolver{path: path, logger: logger.With(zap.String("resolver", "local"))}
}

type localResolver struct {
	sync.Mutex
	path        string
	logger      *zap.Logger
	local       *dns.ClientConfig
	lastModTime time.Time
}

func (r *localResolver) Resolve(msg *dns.Msg) (*dns.Msg, error) {

	local, err := r.loadLocalConfig()

	if err != nil {
		return nil, err
	}

	response := &dns.Msg{}

	name := msg.Question[0].Name
	names := local.NameList(name)
	servers := local.Servers
	r.logger.Debug("LOCAL: Resolving", zap.String("name", name), zap.Strings("names", names))
Outer:
	for _, n := range names {
		for _, s := range servers {
			resolver := NewDns(s, r.logger)
			msg.Question[0].Name = n
			rx, err := resolver.Resolve(msg)
			if err != nil {
				continue
			}

			if rx != nil && len(rx.Answer) > 0 {
				response = rx
				break Outer
			}
		}
		msg.Question[0].Name = name
	}

	response.SetReply(msg)
	return response, nil
}

func (r *localResolver) loadLocalConfig() (*dns.ClientConfig, error) {

	r.Lock()
	defer r.Unlock()

	stat, err := os.Stat(r.path)

	if err != nil {
		return nil, err
	}

	if stat.ModTime().Before(r.lastModTime) {
		return r.local, nil
	}

	r.logger.Debug("LOCAL: Loading local config", zap.String("path", r.path))
	lc, err := dns.ClientConfigFromFile(r.path)

	if err != nil {
		r.logger.Panic("Can't load resolv.conf", zap.Error(err), zap.String("path", r.path))
		return nil, err
	}

	r.local = lc

	return r.local, nil

}
