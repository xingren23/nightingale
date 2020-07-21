package dataobj

import (
	"encoding/json"
	"errors"

	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/logger"
)

type Instance struct {
	Id             int64  `json:"id"`
	AppCode        string `json:"appCode"`
	AppId          int    `json:"appId"`
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

func getInstances(url string, pageNo, pageSize int) (*InstPageResult, error) {
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m InstPageResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Error:Parse Instance JSON %v.", err)
		return nil, err
	}

	return &m, nil
}

// 获取实例
func GetInstByPage() ([]*Instance, error) {
	res := []*Instance{}
	url := config.Get().Api.OpsAddr + "/instance/search"
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getInstances(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page == nil {
			return res, errors.New("page instance result is nil")
		}
		pageNo++
		pageTotal = page.Pagination.TotalPage
		res = append(res, page.Result...)
	}

	return res, nil
}
