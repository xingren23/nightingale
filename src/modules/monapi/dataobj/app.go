package dataobj

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/didi/nightingale/src/modules/monapi/config"
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

func getApps(url string, pageNo, pageSize int) (*AppResult, error) {
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize
	params["sourceType"] = SourceApp

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m AppResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Printf("Error : Cache Instance Parse JSON %v.\n", err)
		return nil, err
	}

	return &m, nil
}

// 获取实例
func GetAppByPage() ([]*App, error) {
	res := []*App{}
	url := config.Get().Api.Ops + "/api/resource/query"
	pageNo := 1
	pageSize := 10
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getApps(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page.Result == nil {
			return res, errors.New("page result is nil")
		}
		pageNo++
		pageTotal = page.Result.Pagination.TotalPage
		res = append(res, page.Result.Result...)
	}

	return res, nil
}
