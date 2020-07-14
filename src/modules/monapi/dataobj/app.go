package dataobj

import (
	"encoding/json"
	"errors"
	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/grafana/grafana/pkg/cmd/grafana-cli/logger"
)

type App struct {
	Id        int64  `json:"id"`
	Code      string `json:"code"`
	Language  string `json:"language"`
	Name      string `json:"name"`
	SrvTreeId int64  `json:"srvTreeId"`
	Basic     bool   `json:"basic"`
}

type AppPageResult struct {
	Pagination Pagination `json:"pagination"`
	Result     []*App     `json:"result"`
}

type AppResult struct {
	Message string         `json:"message"`
	Result  *AppPageResult `json:"result"`
}

func getApps(url string, pageNo, pageSize int) (*AppPageResult, error) {
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m AppPageResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Error: Parse app JSON %v.", err)
		return nil, err
	}

	return &m, nil
}

// 获取应用
func GetAppByPage() ([]*App, error) {
	res := []*App{}
	url := config.Get().Api.Ops + "/app/search"
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getApps(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page == nil {
			return res, errors.New("get app page result is nil")
		}
		pageNo++
		pageTotal = page.Pagination.TotalPage
		res = append(res, page.Result...)
	}

	return res, nil
}
