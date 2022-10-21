package main

import (
	"testing"

	"github.com/shawnburke/dnsmock/config"
	"go.uber.org/fx/fxtest"
)

func TestMain(m *testing.T) {
	params := config.Parameters{}

	graph := buildGraph(params, nil)

	fxtest.New(m, graph)

}
