package main

import (
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeConfig(t *testing.T) {
	config := parseDefaultConfig()

	minimalConfig := `bind = "1.1.1.1:5301" `

	if err := toml.Unmarshal([]byte(minimalConfig), &config); err != nil {
		t.Fatalf("should parse config\n%v", err)
	}

	assert.Equal(t, "1.1.1.1:5301", config.Bind, "expected overridden value for config.bind")
}

func TestMetricsNested(t *testing.T) {
	config := parseDefaultConfig()
	assert.Equal(t, false, config.Metrics.Enabled, "expected metrics off by default")
	assert.Equal(t, "/metrics", config.Metrics.Path, "expected metrics off by default")

	minimalConfig := `

[metrics]
enabled = true
path = "/voo"
`

	if err := toml.Unmarshal([]byte(minimalConfig), &config); err != nil {
		t.Fatalf("should parse config\n%v\n", err)
	}
	assert.Equal(t, true, config.Metrics.Enabled, "expected overridden value for config.bind")
}
