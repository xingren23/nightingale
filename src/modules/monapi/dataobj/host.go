package dataobj

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/didi/nightingale/src/modules/monapi/config"
)

type CmdbHost struct {
	Id             int64  `json:"id"`
	Ip             string `json:"innerIp"`
	HostName       string `json:"hostName"`
	Type           string `json:"type"`
	EnvCode        string `json:"envCode"`
	DataCenterCode string `json:"dataCenterCode"`
}

type CmdbHostPageResult struct {
	Pagination Pagination  `json:"pagination"`
	Result     []*CmdbHost `json:"result"`
}

type CmdbHostResult struct {
	Message string              `json:"message"`
	Result  *CmdbHostPageResult `json:"result"`
}

func getHosts(url string, pageNo, pageSize int) (*CmdbHostResult, error) {
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize
	params["sourceType"] = SourceHost

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m CmdbHostResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Printf("Error : Cache Instance Parse JSON %v.\n", err)
		return nil, err
	}

	return &m, nil
}

// 获取实例
func GetHostByPage() ([]*CmdbHost, error) {
	res := []*CmdbHost{}
	url := config.Get().Api.Ops + "/api/resource/query"
	pageNo := 1
	pageSize := 10
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getHosts(url, pageNo, pageSize)
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
