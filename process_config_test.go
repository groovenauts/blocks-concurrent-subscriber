package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadProcessConfig1(t *testing.T) {
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

func TestLoadProcessConfig2(t *testing.T) {
	config, err := LoadProcessConfig("./test/config2.json")
	assert.NoError(t, err)
	assert.NotEmpty(t, config.Datasource)
	assert.Nil(t, config.Agent)
	assert.NotNil(t, config.Subscriptions)
	assert.Equal(t, 2, len(config.Subscriptions))
	sub0 := config.Subscriptions[0]
	sub1 := config.Subscriptions[1]
	assert.Equal(t, "pipeline1", sub0.Pipeline)
	assert.Equal(t, "pipeline2", sub1.Pipeline)
	assert.Equal(t, "projects/proj-dummy-999/subscriptions/pipeline1-progress-subscription", sub0.Name)
	assert.Equal(t, "projects/proj-dummy-999/subscriptions/pipeline2-progress-subscription", sub1.Name)

	assert.Equal(t, 10, config.Interval)
	if assert.NotNil(t, config.Patterns) {
		assert.Equal(t, 2, len(config.Patterns))
	}
}
