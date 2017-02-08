package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
)

type DefaultAgentClient struct{
	httpClient   *http.Client
	httpUrl      string
	httpToken    string
}

func (ac *DefaultAgentClient) getSubscriptions(ctx context.Context) ([]*Subscription, error) {
	if ac.httpClient == nil {
		ac.httpClient = new(http.Client)
	}

	req, err := http.NewRequest("GET", ac.httpUrl+"/pipelines/subscriptions", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ac.httpToken)
	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

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
