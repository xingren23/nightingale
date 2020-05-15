package dataobj

type InfluxDBItem struct {
	Measurement string                 `json:"metric"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
	Timestamp   int64                  `json:"timestamp"`
}

func (i *InfluxDBItem) PK() string {
	return i.Measurement
}
