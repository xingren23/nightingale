package dataobj

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/didi/nightingale/src/modules/monapi/config"
)

type Instance struct {
	AppCode        string `json:"appCode"`
	AppID          int    `json:"appId"`
	DataCenterCode string `json:"dataCenterCode"`
	EnvCode        string `json:"envCode"`
	GroupCode      string `json:"groupCode"`
	IP             string `json:"ip"`
	HostName       string `json:"hostName"`
	Port           int    `json:"port"`
	UUID           string `json:"uuid"`
}

type InstPageResult struct {
	Pagination Pagination  `json:"pagination"`
	Result     []*Instance `json:"result"`
}

type InstResult struct {
	Message string          `json:"message"`
	Result  *InstPageResult `json:"result"`
}

func getInstances(url string, pageNo, pageSize int) (*InstResult, error) {
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize
	params["sourceType"] = SourceInst

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m InstResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Printf("Error : Cache Instance Parse JSON %v.\n", err)
		return nil, err
	}

	return &m, nil
}

// 获取实例
func GetInstByPage() ([]*Instance, error) {
	res := []*Instance{}
	url := config.Get().Api.Ops + "/api/resource/query"
	pageNo := 1
	pageSize := 10
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getInstances(url, pageNo, pageSize)
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
