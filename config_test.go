package main

import (
	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeConfig(t *testing.T) {
	config := parseDefaultConfig()

	minimalConfig := "bind = \"1.1.1.1:5301\" "

	if _, err := toml.Decode(minimalConfig, &config); err != nil {
		t.Fatalf("should parse config")
	}

	assert.Equal(t, "1.1.1.1:5301", config.Bind, "expected overridden value for config.bind")
}
