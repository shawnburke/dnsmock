package config

type Parameters struct {
	ListenAddr  string   `yaml:"listen_addr"`
	Downstreams []string `yaml:"downstream"`
	Search      []string `yaml:"search"`
}
