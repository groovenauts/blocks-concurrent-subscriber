package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
)

type HttpRequester interface {
	Do(req *http.Request) (*http.Response, error)
}

type InvalidHttpResponse struct {
	StatusCode int
	Msg        string
}

func (e *InvalidHttpResponse) Error() string {
	return fmt.Sprintf("%v %v", e.StatusCode, e.Msg)
}

type DefaultAgentClient struct {
	httpRequester HttpRequester
	httpUrl       string
	httpToken     string
}

func (ac *DefaultAgentClient) getSubscriptions(ctx context.Context) ([]*Subscription, error) {
	url := ac.httpUrl + "/pipelines/subscriptions"

	var subscriptions []Subscription
	err := ac.processRequest(ctx, url, func(body []byte, logAttrs log.Fields) error {
		err := json.Unmarshal(body, &subscriptions)
		if err != nil {
			logAttrs["error"] = err
			log.WithFields(logAttrs).Errorln("Failed to json.Unmarshal")
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := []*Subscription{}
	for _, subscription := range subscriptions {
		result = append(result, &subscription)
	}
	return result, nil
}

func (ac *DefaultAgentClient) processRequest(ctx context.Context, url string, f func([]byte, log.Fields) error) error {
	if ac.httpRequester == nil {
		ac.httpRequester = new(http.Client)
	}
	logAttrs := log.Fields{"url": url}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(logAttrs).Errorln("Failed to http.NewRequest")
		return err
	}
	req.Header.Set("Authorization", "Bearer "+ac.httpToken)

	resp, err := ac.httpRequester.Do(req)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorln("Failed to send HTTP request")
		return err
	}
	defer resp.Body.Close()

	logAttrs["status"] = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = &InvalidHttpResponse{StatusCode: resp.StatusCode, Msg: "Unexpected response"}
		logAttrs["error"] = err
		log.WithFields(logAttrs).Warnln("Server returned error")
		return err
	}

	log.WithFields(logAttrs).Debugln("GET OK")

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorln("Failed to read response body")
		return err
	}

	return f(byteArray, logAttrs)
}
