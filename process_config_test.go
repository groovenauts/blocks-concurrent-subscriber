package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadProcessConfig(t *testing.T) {
	config, err := LoadProcessConfig("./test/config1.json")
	assert.NoError(t, err)

	assert.NotEmpty(t, config.Datasource)
	assert.NotNil(t, config.Agent)
	assert.NotEmpty(t, config.Agent.RootUrl)
	assert.NotEmpty(t, config.Agent.Token)
	assert.Equal(t, 10, config.Interval)
	if assert.NotNil(t, config.Patterns) {
		assert.Equal(t, 2, len(config.Patterns))
	}
}
