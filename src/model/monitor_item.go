package model

import (
	"time"
)

const (
	EndpointTypeInstance = "INSTANCE"
	EndpointTypeHost     = "HOST"
	EndpointTypeNetwork  = "NETWORK"
)

// TODO : 去掉不用的字段
type MonitorItem struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name"`
	Metric       string    `json:"metric"`
	Type         string    `json:"type"`
	Category     string    `json:"category"`
	EndpointType string    `json:"endpoint_type"`
	Step         int       `json:"step"`
	Unit         string    `json:"unit"`
	InfluxType   string    `json:"influx_type"`
	Measurement  string    `json:"measurement"`
	MachineType  string    `json:"machine_type"`
	Description  string    `json:"description"`
	CreateTime   time.Time `json:"create_time"`
	CreateBy     int       `json:"create_by"`
	UpdateTime   time.Time `json:"update_time"`
	UpdateBy     int       `json:"update_by"`
	Status       int       `json:"status"`
}

func MonitorItemAll() ([]*MonitorItem, error) {
	objs := make([]*MonitorItem, 0)

	err := DB["mon"].Where("status > -1").Find(&objs)
	if err != nil {
		return objs, err
	}

	items := make([]*MonitorItem, 0)
	for _, obj := range objs {
		items = append(items, obj)
	}
	return items, err
}
