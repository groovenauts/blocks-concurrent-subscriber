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
		config: &AgentConfig{
			RootUrl: "http://somewhere",
			Token:   "DUMMY-TOKEN",
		},
	}
	ctx := context.Background()
	subscriptions, err := ac.getSubscriptions(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(subscriptions))
	sub := subscriptions[0]
	assert.Equal(t, "pipeline01", sub.Pipeline)
	assert.Equal(t, "pipeline01-progress-subscription", sub.Name)
}

func TestConvertToSubscriptionPointerArray(t *testing.T) {
	ctx := context.Background()
	s1 := Subscription{ Name: "sub1"}
	s2 := Subscription{ Name: "sub2" }
	s3 := Subscription{ Name: "sub3" }
	subscriptions := []Subscription{s1, s2, s3}

	ac := &DefaultAgentClient{}

	res := ac.convertToSubscriptionPointerArray(ctx, subscriptions)
	assert.Equal(t, len(subscriptions), len(res))

	assert.Equal(t, "sub1", res[0].Name)
	assert.Equal(t, "sub2", res[1].Name)
	assert.Equal(t, "sub3", res[2].Name)
}
