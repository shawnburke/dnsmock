package config

import (
	"strings"
	"time"
)

type Parameters struct {
	ListenAddr     string        `yaml:"listen_addr"`
	DownstreamsRaw string        `yaml:"downstreams"`
	Search         []string      `yaml:"search"`
	RecordTTL      time.Duration `yaml:"record_ttl"`
	Record         bool          `yaml:"record"`
	ReplayFile     string        `yaml:"replay_file"`
}

func (p Parameters) Downstreams() []string {
	parts := strings.Split(p.DownstreamsRaw, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

func (p Parameters) IsDownstreamLocalhost() bool {
	ds := p.Downstreams()
	return len(ds) == 1 && ds[0] == "localhost"
}
