package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
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

	url := ac.httpUrl+"/pipelines/subscriptions"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ac.httpToken)

	resp, err := ac.httpRequester.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"url": url}).Errorln(err)
		return nil, err
	}
	defer resp.Body.Close()

	log.WithFields(log.Fields{"url": url, "status": resp.StatusCode}).Debugln("GET OK")

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var subscriptions []Subscription
	err = json.Unmarshal(byteArray, &subscriptions)
	if err != nil {
		log.WithFields(log.Fields{"url": url}).Errorln(err)
		return nil, err
	}
	result := []*Subscription{}
	for _, subscription := range subscriptions {
		result = append(result, &subscription)
	}
	return result, nil
}
