package dataobj

import (
	"encoding/json"
	"log"

	"github.com/didi/nightingale/src/modules/monapi/config"
)

type MonitorItem struct {
	Id           int64  `json:"id"`
	EndpointType string `json:"endpointType"`
	Metric       string `json:"metric"`
	Name         string `json:"name"`
	Step         int    `json:"step"`
	Type         string `json:"type"`
	Tags         string `json:"tags"`
	Unit         string `json:"unit"`
	InfluxType   string `json:"influxType"`
	Measurement  string `json:"measurement"`
	Field        string `json:"Field"`
}

type MetricPageResult struct {
	Pagination Pagination     `json:"pagination"`
	Result     []*MonitorItem `json:"result"`
}

func getMonitorItems(url string, pageNo, pageSize int) (*MetricPageResult, error) {
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}
	var m MetricPageResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Printf("Error : Cache metrics Parse JSON %v.\n", err)
		return nil, err
	}

	return &m, nil
}

// 获取指标元数据
func GetMetricByPage() ([]*MonitorItem, error) {
	res := []*MonitorItem{}
	url := config.Get().Api.Ops + "/metric/info"
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getMonitorItems(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		pageNo++
		pageTotal = page.Pagination.TotalPage
		res = append(res, page.Result...)
	}

	return res, nil
}
