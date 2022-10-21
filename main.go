package main

import (
	"context"
	"fmt"
	"time"

	"flag"

	"github.com/shawnburke/dnsmock/config"
	"github.com/shawnburke/dnsmock/proxy"
	"github.com/shawnburke/dnsmock/spec"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {

	cfg := config.Parameters{
		ListenAddr:     "0.0.0.0:50053",
		RecordTTL:      time.Minute,
		DownstreamsRaw: "8.8.8.8:53",
	}

	flag.BoolVar(&cfg.Record, "record", false, "Record responses")
	flag.StringVar(&cfg.ReplayFile, "replay-file", "", "Replay from file")
	flag.StringVar(&cfg.DownstreamsRaw, "downstreams", "localhost", "Downstreams, comma separated")

	flag.Parse()

	graph := buildGraph(cfg, func(ctx context.Context, s spec.Responses) {
		if cfg.Record {
			result := s.YAML()
			fmt.Println(result)
		}
	})

	fx.New(graph).Run()

}

func buildGraph(cfg config.Parameters, shutdown func(ctx context.Context, s spec.Responses)) fx.Option {
	return fx.Options(
		fx.Provide(
			func() config.Parameters {
				return cfg
			},
			func(cfg config.Parameters) spec.Responses {
				if cfg.ReplayFile == "" {
					return spec.New()
				}
				return spec.FromFile(cfg.ReplayFile)
			},
			zap.NewDevelopment,
			proxy.NewForwarder,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, s spec.Responses, cfg config.Parameters, logger *zap.Logger) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						logger.Info("Config", zap.Any("cfg", cfg))
						return nil
					},
					OnStop: func(ctx context.Context) error {
						if shutdown != nil {
							shutdown(ctx, s)
						}
						return nil
					},
				})
			},
			proxy.New,
		),
	)
}
