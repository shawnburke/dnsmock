package proxy

import (
	"testing"

	"github.com/shawnburke/dnsmock/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestProxy(t *testing.T) {
	logger := zap.NewNop()
	p := New(logger, config.Parameters{
		ListenAddr: "0.0.0.0:0",
	}, nil)
	require.NotNil(t, p)

	err := p.Start()
	require.NoError(t, err)
	p.Stop()

}
