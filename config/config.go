package config

import (
	"fmt"
	"strings"
)

type Parameters struct {
	Port           int    `yaml:"port"`
	DownstreamsRaw string `yaml:"downstreams"`
	Record         bool   `yaml:"record"`
	ReplayFile     string `yaml:"replay_file"`
	RecordFile     string `yaml:"record_file"`
	Verbose        bool   `yaml:"verbose"`
}

func (p Parameters) ListenAddr() string {
	return fmt.Sprintf(":%d", p.Port)
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
