package backend

import (
	"time"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/toolkits/pkg/concurrent/semaphore"
	"github.com/toolkits/pkg/container/list"
	"github.com/toolkits/pkg/logger"
)

type InfluxDBStorage struct {
	// config
	section InfluxdbSection

	// 发送缓存队列 node -> queue_of_data
	InfluxdbQueue *list.SafeListLimited
}

func (influx *InfluxDBStorage) Init() {

	// init queue
	if influx.section.Enabled {
		influx.InfluxdbQueue = list.NewSafeListLimited(DefaultSendQueueMaxSize)
	}

	// init task
	influxdbConcurrent := influx.section.WorkerNum
	if influxdbConcurrent < 1 {
		influxdbConcurrent = 1
	}
	go influx.send2InfluxdbTask(influxdbConcurrent)

	// TODO
	RegisterPushEndpoint(influx.section.Name, influx)
}

// TODO 实现 query 接口
//func (influx *InfluxDBStorage) QueryData(inputs []dataobj.QueryData) []*dataobj.TsdbQueryResponse {
//	return nil
//}

// 将原始数据插入到influxdb缓存队列
func (influx *InfluxDBStorage) Push2Queue(items []*dataobj.MetricValue) {
	errCnt := 0
	for _, item := range items {
		influxdbItem := influx.convert2InfluxdbItem(item)
		isSuccess := influx.InfluxdbQueue.PushFront(influxdbItem)

		if !isSuccess {
			errCnt += 1
		}
	}
	stats.Counter.Set("influxDB.queue.err", errCnt)
}

func (influx *InfluxDBStorage) send2InfluxdbTask(concurrent int) {
	batch := influx.section.Batch // 一次发送,最多batch条数据
	retry := influx.section.MaxRetry
	addr := influx.section.Address
	sema := semaphore.NewSemaphore(concurrent)

	var err error
	c, err := NewInfluxdbClient(influx.section)
	defer c.Client.Close()

	if err != nil {
		logger.Errorf("init influxDB client fail: %v", err)
		return
	}

	for {
		items := influx.InfluxdbQueue.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		influxdbItems := make([]*dataobj.InfluxdbItem, count)
		for i := 0; i < count; i++ {
			influxdbItems[i] = items[i].(*dataobj.InfluxdbItem)
			stats.Counter.Set("points.out.influxDB", 1)
			logger.Debug("send to influxDB: ", influxdbItems[i])
		}

		//  同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(addr string, influxdbItems []*dataobj.InfluxdbItem, count int) {
			defer sema.Release()
			sendOk := false

			for i := 0; i < retry; i++ {
				err = c.Send(influxdbItems)
				if err == nil {
					sendOk = true
					break
				}
				logger.Warningf("send influxDB fail: %v", err)
				time.Sleep(time.Millisecond * 10)
			}

			if !sendOk {
				stats.Counter.Set("points.out.influxDB.err", count)
				logger.Errorf("send %v to influxDB %s fail: %v", influxdbItems, addr, err)
			} else {
				logger.Debugf("send to influxDB %s ok", addr)
			}
		}(addr, influxdbItems, count)
	}
}

func (influx *InfluxDBStorage) convert2InfluxdbItem(d *dataobj.MetricValue) *dataobj.InfluxdbItem {
	t := dataobj.InfluxdbItem{Tags: make(map[string]string), Fields: make(map[string]interface{})}

	for k, v := range d.TagsMap {
		t.Tags[k] = v
	}
	t.Tags["endpoint"] = d.Endpoint
	t.Measurement = d.Metric
	t.Fields["value"] = d.Value
	t.Timestamp = d.Timestamp

	return &t
}

type InfluxClient struct {
	Client    client.Client
	Database  string
	Precision string
}

func NewInfluxdbClient(section InfluxdbSection) (*InfluxClient, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     section.Address,
		Username: section.Username,
		Password: section.Password,
		Timeout:  time.Millisecond * time.Duration(section.Timeout),
	})

	if err != nil {
		return nil, err
	}

	return &InfluxClient{
		Client:    c,
		Database:  section.Database,
		Precision: section.Precision,
	}, nil
}

func (c *InfluxClient) Send(items []*dataobj.InfluxdbItem) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  c.Database,
		Precision: c.Precision,
	})
	if err != nil {
		logger.Error("create batch points error: ", err)
		return err
	}

	for _, item := range items {
		pt, err := client.NewPoint(item.Measurement, item.Tags, item.Fields, time.Unix(item.Timestamp, 0))
		if err != nil {
			logger.Error("create new points error: ", err)
			continue
		}
		bp.AddPoint(pt)
	}

	return c.Client.Write(bp)
}
