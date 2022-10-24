package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"flag"

	"github.com/shawnburke/dnsmock/config"
	"github.com/shawnburke/dnsmock/proxy"
	"github.com/shawnburke/dnsmock/resolver"
	"github.com/shawnburke/dnsmock/spec"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {

	cfg := config.Parameters{
		ListenAddr:     "0.0.0.0:50053",
		DownstreamsRaw: "",
	}

	flag.BoolVar(&cfg.Record, "record", false, "Record responses")
	flag.StringVar(&cfg.ReplayFile, "replay-file", "", "Replay from file")
	flag.StringVar(&cfg.RecordFile, "record-file", "", "Record to file")
	flag.StringVar(&cfg.ListenAddr, "listen-addr", cfg.ListenAddr, "Listen address")
	flag.StringVar(&cfg.DownstreamsRaw, "downstreams", resolver.DownstreamLocalhost, "Downstreams, comma separated or 'none' to prevent downstream lookup")

	flag.Parse()

	graph := buildGraph(cfg, func(ctx context.Context, s *spec.Responses) {
		if cfg.Record {
			result := s.YAML()

			if cfg.RecordFile != "" {
				err := ioutil.WriteFile(cfg.RecordFile, []byte(result), 0644)
				if err != nil {
					fmt.Printf("Error writing to file %q: %v", cfg.RecordFile, err.Error())
				}
			} else {
				fmt.Println(result)
			}
		}
	})

	fx.New(graph).Run()

}

func buildGraph(cfg config.Parameters, shutdown func(ctx context.Context, s *spec.Responses)) fx.Option {
	return fx.Options(
		fx.Provide(
			func() config.Parameters {
				return cfg
			},
			func(cfg config.Parameters) *spec.Responses {
				if cfg.ReplayFile != "" {
					return spec.FromFile(cfg.ReplayFile)
				}

				if cfg.Record || cfg.RecordFile != "" {
					return spec.New()
				}
				return nil
			},
			zap.NewDevelopment,
			resolver.Build,
			proxy.NewFromConfig,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, p proxy.Interface, s *spec.Responses, cfg config.Parameters, logger *zap.Logger) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						logger.Info("Config", zap.Any("cfg", cfg))
						return p.Start()
					},
					OnStop: func(ctx context.Context) error {
						if shutdown != nil {
							shutdown(ctx, s)
						}
						return p.Stop()
					},
				})
			},
		),
	)
}
