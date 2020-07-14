package dataobj

import (
	"encoding/json"
	"errors"
	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/grafana/grafana/pkg/cmd/grafana-cli/logger"
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

type NetworkPageResult struct {
	Pagination Pagination `json:"pagination"`
	Networks   []*Network `json:"result"`
}

type NetworkResult struct {
	Message string             `json:"message"`
	Result  *NetworkPageResult `json:"result"`
}

func getNetworks(url string, pageNo, pageSize int) (*NetworkResult, error) {
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize
	params["sourceType"] = CmdbSourceNet

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m NetworkResult
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
	url := config.Get().Api.Ops + "/api/resource/query"
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getNetworks(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page.Result == nil {
			return res, errors.New("page result is nil")
		}
		pageNo++
		pageTotal = page.Result.Pagination.TotalPage
		res = append(res, page.Result.Networks...)
	}

	return res, nil
}
