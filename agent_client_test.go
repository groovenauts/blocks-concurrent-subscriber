package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
)

type DummyHttp struct{}

func (dh *DummyHttp) Do(req *http.Request) (*http.Response, error) {
	resp := `[{"pipeline":"pipeline01","subscription":"pipeline01-progress-subscription"}]`
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(resp)),
	}, nil
}

func TestGetSubscriptions(t *testing.T) {
	ac := &DefaultAgentClient{
		httpRequester: &DummyHttp{},
		httpUrl:       "http://somewhere",
		httpToken:     "DUMMY-TOKEN",
	}
	ctx := context.Background()
	subscriptions, err := ac.getSubscriptions(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(subscriptions))
	sub := subscriptions[0]
	assert.Equal(t, "pipeline01", sub.Pipeline)
	assert.Equal(t, "pipeline01-progress-subscription", sub.Name)
}
