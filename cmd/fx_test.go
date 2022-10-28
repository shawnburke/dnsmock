package main

import (
	"testing"

	"github.com/shawnburke/dnsmock/config"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestMain(m *testing.T) {
	params := config.Parameters{}

	graph := buildGraph(params, zap.NewNop(), nil)

	fxtest.New(m, graph)

}
