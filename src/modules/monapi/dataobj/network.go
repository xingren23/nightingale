package dataobj

import (
	"encoding/json"
	"errors"

	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/logger"
)

type Network struct {
	Id             int64  `json:"id"`
	ManageIp       string `json:"manageIp"`
	Name           string `json:"name"`
	DataCenterCode string `json:"dataCenterCode"`
	EnvCode        string `json:"envCode"`
	Type           string `json:"type"`
	SerialNo       string `json:"serialNo"`
}

type NetworkSearchResult struct {
	Message    string     `json:"message"`
	Pagination Pagination `json:"pagination"`
	Result     []*Network `json:"result"`
}

func getNetworks(url string, pageNo, pageSize int) (*NetworkSearchResult, error) {
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m NetworkSearchResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Error: Parse network JSON %v.", err)
		return nil, err
	}

	return &m, nil
}

// 获取实例
func GetNetByPage() ([]*Network, error) {
	res := []*Network{}
	url := config.Get().Api.OpsAddr + "/network_device/search"
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getNetworks(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page == nil {
			return res, errors.New("page result is nil")
		}
		pageNo++
		pageTotal = page.Pagination.TotalPage
		res = append(res, page.Result...)
	}

	return res, nil
}
