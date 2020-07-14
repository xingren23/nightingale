package dataobj

import (
	"encoding/json"
	"errors"
	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/logger"
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
	Hosts      []*CmdbHost `json:"result"`
}

type CmdbHostResult struct {
	Message string              `json:"message"`
	Result  *CmdbHostPageResult `json:"result"`
}

func getHosts(url string, pageNo, pageSize int) (*CmdbHostPageResult, error) {
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m CmdbHostPageResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Error: Parse cmdbHost JSON %v.", err)
		return nil, err
	}

	return &m, nil
}

// 获取实例
func GetHostByPage() ([]*CmdbHost, error) {
	res := []*CmdbHost{}
	url := config.Get().Api.Ops + "/host/search"
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getHosts(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page == nil {
			return res, errors.New("get host page result is nil")
		}
		pageNo++
		pageTotal = page.Pagination.TotalPage
		res = append(res, page.Hosts...)
	}

	return res, nil
}
