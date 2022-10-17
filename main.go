package main

import (
	"github.com/shawnburke/dnsmock/config"
	"github.com/shawnburke/dnsmock/proxy"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {

	cfg := config.Parameters{
		ListenAddr:  "0.0.0.0:50053",
		Downstreams: []string{"8.8.8.8:53"},
	}

	graph := buildGraph(cfg)

	fx.New(graph).Run()

}

func buildGraph(cfg config.Parameters) fx.Option {
	return fx.Options(
		fx.Provide(func() config.Parameters {
			return cfg
		}),
		fx.Provide(
			zap.NewDevelopment,
		),
		fx.Invoke(
			proxy.New,
		),
	)
}
