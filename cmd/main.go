package main

import (
	"context"
	"fmt"
	"os"

	"flag"

	"github.com/shawnburke/dnsmock"
	"github.com/shawnburke/dnsmock/config"
	"github.com/shawnburke/dnsmock/resolver"
	"github.com/shawnburke/dnsmock/spec"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

const defaultPort = 53

func main() {

	cfg := config.Parameters{
		Port:           defaultPort,
		DownstreamsRaw: "",
	}

	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose logging")
	flag.BoolVar(&cfg.Record, "record", false, "Record responses")
	flag.StringVar(&cfg.ReplayFile, "replay-file", "", "Replay from file")
	flag.StringVar(&cfg.RecordFile, "record-file", "", "Record to file")
	flag.IntVar(&cfg.Port, "port", defaultPort, "Listen port")
	flag.StringVar(&cfg.DownstreamsRaw, "downstreams", resolver.DownstreamLocalhost, "Downstreams, comma separated or 'none' to prevent downstream lookup")

	flag.Parse()

	logger, err := buildLogger(cfg)
	if err != nil {
		panic(err)
	}

	graph := buildGraph(cfg, logger, func(ctx context.Context, s *spec.Responses) {
		if cfg.Record {
			result := s.YAML()

			if cfg.RecordFile != "" {
				err := os.WriteFile(cfg.RecordFile, []byte(result), 0644)
				if err != nil {
					fmt.Printf("Error writing to file %q: %v", cfg.RecordFile, err.Error())
				}
			} else {
				fmt.Println(result)
			}
		}
	})

	fx.New(fx.WithLogger(func() fxevent.Logger {
		return &fxevent.ZapLogger{Logger: logger}
	}), graph).Run()

}

func buildLogger(cfg config.Parameters) (*zap.Logger, error) {
	zapcfg := zap.NewDevelopmentConfig()

	if !cfg.Verbose {
		zapcfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	}
	return zapcfg.Build()
}

func buildSpecResponses(cfg config.Parameters) *spec.Responses {
	if cfg.ReplayFile != "" {
		return spec.FromFile(cfg.ReplayFile)
	}

	if cfg.Record || cfg.RecordFile != "" {
		return spec.New()
	}
	return nil
}

func buildGraph(cfg config.Parameters,
	logger *zap.Logger,
	shutdown func(ctx context.Context, s *spec.Responses)) fx.Option {
	return fx.Options(
		fx.Supply(logger),
		fx.Supply(cfg),
		fx.Provide(
			buildSpecResponses,
			resolver.Build,
			dnsmock.NewFromConfig,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, p dnsmock.Proxy, s *spec.Responses, cfg config.Parameters, logger *zap.Logger) {
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
