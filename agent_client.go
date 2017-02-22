package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
)

type HttpRequester interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultAgentClient struct {
	httpRequester HttpRequester
	httpUrl       string
	httpToken     string
}

func (ac *DefaultAgentClient) getSubscriptions(ctx context.Context) ([]*Subscription, error) {
	if ac.httpRequester == nil {
		ac.httpRequester = new(http.Client)
	}

	req, err := http.NewRequest("GET", ac.httpUrl+"/pipelines/subscriptions", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ac.httpToken)
	resp, err := ac.httpRequester.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("DefaultAgentClient.getSubscriptions()\n%v\n", string(byteArray))

	var subscriptions []Subscription
	err = json.Unmarshal(byteArray, &subscriptions)
	if err != nil {
		return nil, err
	}
	result := []*Subscription{}
	for _, subscription := range subscriptions {
		result = append(result, &subscription)
	}
	return result, nil
}
