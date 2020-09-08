package model

import (
	"time"
)

const (
	EndpointTypeInstance = "INSTANCE"
	EndpointTypePm       = "PM"
	EndpointTypeDocker   = "DOCKER"
	EndpointTypeNetwork  = "NETWORK"
)

type MetricInfo struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name"`
	Metric       string    `json:"metric"`
	Type         string    `json:"type"`
	Step         int       `json:"step"`
	Unit         string    `json:"unit"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	EndpointType string    `json:"endpoint_type"`
	MachineType  string    `json:"machine_type"`
	CreateTime   time.Time `json:"create_time"`
	CreateBy     int       `json:"create_by"`
	UpdateTime   time.Time `json:"update_time"`
	UpdateBy     int       `json:"update_by"`
	Status       int       `json:"status"`
}

func MetricInfoAll() ([]*MetricInfo, error) {
	objs := make([]*MetricInfo, 0)

	err := DB["mon"].Where("status > -1").Find(&objs)
	if err != nil {
		return objs, err
	}

	items := make([]*MetricInfo, 0)
	for _, obj := range objs {
		items = append(items, obj)
	}
	return items, err
}
