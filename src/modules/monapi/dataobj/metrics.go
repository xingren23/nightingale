package dataobj

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
