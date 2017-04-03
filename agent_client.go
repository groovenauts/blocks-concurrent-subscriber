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

type InvalidPipeline struct {
	Msg        string
}

func (e *InvalidPipeline) Error() string {
	return e.Msg
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
		subscription.isOpened = func() (bool, error) {
			st, err := ac.getPipelineStatus(ctx, subscription.PipelineID)
			if err != nil {
				return false, err
			}
			return st == 4, nil
		}
		result = append(result, &subscription)
	}
	return result, nil
}

func (ac *DefaultAgentClient) getPipelineStatus(ctx context.Context, id string) (int, error) {
	url := ac.httpUrl + "/pipelines/" + id
	var obj map[string]interface{}
	var res int;
	err := ac.processRequest(ctx, url, func(body []byte, logAttrs log.Fields) error {
		err := json.Unmarshal(body, &obj)
		if err != nil {
			logAttrs["error"] = err
			log.WithFields(logAttrs).Errorln("Failed to json.Unmarshal")
			return err
		}
		st, ok := obj["status"]
		if !ok {
			return &InvalidPipeline{Msg: "Response has no status"}
		}
		// json.Unmarshal uses float64 for JSON numbers
		// https://golang.org/pkg/encoding/json/#Unmarshal
		switch st.(type) {
		case float64:
			res = int(st.(float64))
			return nil
		default:
			return &InvalidPipeline{
				Msg: fmt.Sprintf("Status is expected a float64 but it was [%T] %v", st, st),
			}
		}
	})
	if err != nil {
		return 0, err
	}
	return res, nil
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
