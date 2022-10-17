package main

import (
	"fmt"
	"testing"

	"github.com/shawnburke/dnsmock/config"
)

func TestMain(m *testing.T) {
	params := config.Parameters{}

	graph := buildGraph(params)

	fmt.Println(graph.String())

}
