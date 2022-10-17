package config

import "time"

type Parameters struct {
	ListenAddr  string        `yaml:"listen_addr"`
	Downstreams []string      `yaml:"downstream"`
	Search      []string      `yaml:"search"`
	RecordTTL   time.Duration `yaml:"record_ttl"`
}
