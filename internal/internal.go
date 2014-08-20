package internal

import (
	"fmt"
	"net/http"

	"github.com/jmcvetta/napping"
)

type InternalAPI struct {
	Config *InternalAPIConfig
}

type InternalAPIConfig map[string]*InternalAPIEnvironment

type InternalAPIEnvironment struct {
	BaseURL string `json:"base_url"`
	AuthKey string `json:"auth_key"`
}

func New(confLoader func(config interface{}) error) *InternalAPI {
	var conf struct {
		PlotlyInternalEndpoint *InternalAPIConfig
	}
	confLoader(&conf)
	intApi := InternalAPI{
		Config: conf.PlotlyInternalEndpoint,
	}
	return &intApi
}

func (int *InternalAPI) GetCurrentHead(env string) string {
	conf := (*int.Config)[env]
	if conf == nil {
		return ""
	}

	url := conf.BaseURL
	authKey := conf.AuthKey

	if url == "" || authKey == "" {
		return ""
	}

	result := struct {
		CurrentHead string `json:"current_head"`
	}{}

	req := napping.Request{
		Url:    fmt.Sprintf("%s%s", url, "/current_head"),
		Method: "GET",
		Result: &result,
		Header: &http.Header{},
	}
	req.Header.Set("X-Internal-Key", authKey)
	_, err := napping.Send(&req)

	if err != nil {
		return ""
	}

	return result.CurrentHead
}
