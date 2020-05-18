package backend

import (
	"strings"
	"time"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/transfer/cache"
	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/didi/nightingale/src/toolkits/str"
	"github.com/toolkits/pkg/concurrent/semaphore"
	"github.com/toolkits/pkg/container/list"
	"github.com/toolkits/pkg/logger"
)

// send
const (
	DefaultSendTaskSleepInterval = time.Millisecond * 50 //默认睡眠间隔为50ms
	MaxSendRetry                 = 10
)

var (
	MinStep int //最小上报周期,单位sec
)

func startSendTasks() {

	tsdbConcurrent := Config.WorkerNum
	if tsdbConcurrent < 1 {
		tsdbConcurrent = 1
	}

	judgeConcurrent := Config.WorkerNum
	if judgeConcurrent < 1 {
		judgeConcurrent = 1
	}

	influxdbConcurrent := Config.Influxdb.WorkerNum
	if influxdbConcurrent < 1 {
		influxdbConcurrent = 1
	}

	if Config.Enabled {
		judgeQueue := JudgeQueues.GetAll()
		for instance, queue := range judgeQueue {
			go Send2JudgeTask(queue, instance, judgeConcurrent)
		}

		if Config.Tsdb.Enabled {
			for node, item := range Config.Tsdb.ClusterList {
				for _, addr := range item.Addrs {
					queue := TsdbQueues[node+addr]
					go Send2TsdbTask(queue, node, addr, tsdbConcurrent)
				}
			}
		}

		if Config.Influxdb.Enabled {
			for node, item := range Config.Influxdb.ClusterList {
				for _, addr := range item.Addrs {
					queue := InfluxdbQueue[node+addr]
					go send2InfluxDBTask(queue, node, addr, influxdbConcurrent)
				}
			}
		}
	}

}

func Send2TsdbTask(Q *list.SafeListLimited, node, addr string, concurrent int) {
	batch := Config.Batch // 一次发送,最多batch条数据
	sema := semaphore.NewSemaphore(concurrent)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		tsdbItems := make([]*dataobj.TsdbItem, count)
		stats.Counter.Set("points.out.tsdb", count)
		for i := 0; i < count; i++ {
			tsdbItems[i] = items[i].(*dataobj.TsdbItem)
			logger.Debug("send to tsdb->: ", tsdbItems[i])
		}

		//控制并发
		sema.Acquire()
		go func(addr string, tsdbItems []*dataobj.TsdbItem, count int) {
			defer sema.Release()

			resp := &dataobj.SimpleRpcResponse{}
			var err error
			sendOk := false
			for i := 0; i < 3; i++ { //最多重试3次
				err = TsdbConnPools.Call(addr, "Tsdb.Send", tsdbItems, resp)
				if err == nil {
					sendOk = true
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			if !sendOk {
				stats.Counter.Set("points.out.tsdb.err", count)
				logger.Errorf("send %v to tsdb %s:%s fail: %v", tsdbItems, node, addr, err)
			} else {
				logger.Debugf("send to tsdb %s:%s ok", node, addr)
			}
		}(addr, tsdbItems, count)
	}
}

// Push2TsdbSendQueue pushes data to a TSDB instance which depends on the consistent ring.
func Push2TsdbSendQueue(items []*dataobj.MetricValue) {
	errCnt := 0
	for _, item := range items {
		tsdbItem := convert2TsdbItem(item)
		stats.Counter.Set("tsdb.queue.push", 1)

		node, err := TsdbNodeRing.GetNode(item.PK())
		if err != nil {
			logger.Warningf("get tsdb node error: %v", err)
			continue
		}

		cnode := Config.Tsdb.ClusterList[node]
		for _, addr := range cnode.Addrs {
			Q := TsdbQueues[node+addr]
			// 队列已满
			if !Q.PushFront(tsdbItem) {
				errCnt += 1
			}
		}
	}

	// statistics
	if errCnt > 0 {
		stats.Counter.Set("tsdb.queue.err", errCnt)
		logger.Error("Push2TsdbSendQueue err num: ", errCnt)
	}
}

func Send2JudgeTask(Q *list.SafeListLimited, addr string, concurrent int) {
	batch := Config.Batch
	sema := semaphore.NewSemaphore(concurrent)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}
		judgeItems := make([]*dataobj.JudgeItem, count)
		stats.Counter.Set("points.out.judge", count)
		for i := 0; i < count; i++ {
			judgeItems[i] = items[i].(*dataobj.JudgeItem)
			logger.Debug("send to judge: ", judgeItems[i])
		}

		sema.Acquire()
		go func(addr string, judgeItems []*dataobj.JudgeItem, count int) {
			defer sema.Release()

			resp := &dataobj.SimpleRpcResponse{}
			var err error
			sendOk := false
			for i := 0; i < MaxSendRetry; i++ {
				err = JudgeConnPools.Call(addr, "Judge.Send", judgeItems, resp)
				if err == nil {
					sendOk = true
					break
				}
				logger.Warningf("send judge %s fail: %v", addr, err)
				time.Sleep(time.Millisecond * 10)
			}

			if !sendOk {
				stats.Counter.Set("points.out.judge.err", count)
				for _, item := range judgeItems {
					logger.Errorf("send %v to judge %s fail: %v", item, addr, err)
				}
			}

		}(addr, judgeItems, count)
	}
}

func Push2JudgeSendQueue(items []*dataobj.MetricValue) {
	errCnt := 0
	for _, item := range items {
		key := str.PK(item.Metric, item.Endpoint)
		stras := cache.StraMap.GetByKey(key)

		for _, stra := range stras {
			if !TagMatch(stra.Tags, item.TagsMap) {
				continue
			}
			judgeItem := &dataobj.JudgeItem{
				Endpoint:  item.Endpoint,
				Metric:    item.Metric,
				Value:     item.Value,
				Timestamp: item.Timestamp,
				DsType:    item.CounterType,
				Tags:      item.Tags,
				TagsMap:   item.TagsMap,
				Step:      int(item.Step),
				Sid:       stra.Id,
				Extra:     item.Extra,
			}

			q, exists := JudgeQueues.Get(stra.JudgeInstance)
			if exists {
				if !q.PushFront(judgeItem) {
					errCnt += 1
				}
			}
		}
	}

	if errCnt > 0 {
		stats.Counter.Set("judge.queue.err", errCnt)
		logger.Error("Push2JudgeSendQueue err num: ", errCnt)
	}
}

// 打到 Tsdb 的数据,要根据 rrdtool 的特定 来限制 step、counterType、timestamp
func convert2TsdbItem(d *dataobj.MetricValue) *dataobj.TsdbItem {
	item := &dataobj.TsdbItem{
		Endpoint:  d.Endpoint,
		Metric:    d.Metric,
		Value:     d.Value,
		Timestamp: d.Timestamp,
		Tags:      d.Tags,
		TagsMap:   d.TagsMap,
		Step:      int(d.Step),
		Heartbeat: int(d.Step) * 2,
		DsType:    dataobj.GAUGE,
		Min:       "U",
		Max:       "U",
	}

	item.Timestamp = alignTs(item.Timestamp, int64(item.Step))

	return item
}

func alignTs(ts int64, period int64) int64 {
	return ts - ts%period
}

func TagMatch(straTags []model.Tag, tag map[string]string) bool {
	for _, stag := range straTags {
		if _, exists := tag[stag.Tkey]; !exists {
			return false
		}
		var match bool
		if stag.Topt == "=" { //当前策略 tagkey 对应的 tagv
			for _, v := range stag.Tval {
				if tag[stag.Tkey] == v {
					match = true
					break
				}
			}
		} else {
			match = true
			for _, v := range stag.Tval {
				if tag[stag.Tkey] == v {
					match = false
					return match
				}
			}
		}

		if !match {
			return false
		}
	}
	return true
}

// 将原始数据插入到influxdb缓存队列
func Push2InfluxDBSendQueue(items []*dataobj.MetricValue) {
	errCnt := 0
	for _, item := range items {
		influxDBItem := convert2InfluxDBItem(item)

		node, err := InfluxNodeRing.GetNode(influxDBItem.PK())
		if err != nil {
			logger.Warningf("get influxdb node error: %v", err)
			continue
		}

		cnode := Config.Influxdb.ClusterList[node]
		for _, addr := range cnode.Addrs {
			Q := InfluxdbQueue[node+addr]
			// 队列已满
			if !Q.PushFront(influxDBItem) {
				errCnt += 1
			}
		}
	}

	if errCnt > 0 {
		stats.Counter.Set("influxdb.queue.err", errCnt)
		logger.Error("Push2InfluxDBSendQueue err num: ", errCnt)
	}

}

func convert2InfluxDBItem(d *dataobj.MetricValue) *dataobj.InfluxDBItem {
	t := dataobj.InfluxDBItem{Tags: make(map[string]string), Fields: make(map[string]interface{})}

	for k, v := range d.TagsMap {
		t.Tags[k] = v
	}
	t.Tags["endpoint"] = d.Endpoint
	t.Measurement = d.Metric
	if d.CounterType == dataobj.GAUGE {
		t.Fields[strings.ToLower(dataobj.GAUGE)] = d.Value
	} else if d.CounterType == dataobj.COUNTER {
		t.Fields[strings.ToLower(dataobj.COUNTER)] = d.Value
	} else {
		t.Fields["value"] = d.Value
	}
	t.Timestamp = d.Timestamp

	return &t
}

func send2InfluxDBTask(Q *list.SafeListLimited, node, addr string, concurrent int) {
	batch := Config.Influxdb.Batch // 一次发送,最多batch条数据
	retry := Config.Influxdb.MaxRetry
	sema := semaphore.NewSemaphore(concurrent)

	var err error
	influxClient, err := NewInfluxClient(addr)
	if err != nil {
		logger.Errorf("init influxdb client fail: %v", err)
		return
	}
	defer influxClient.Client.Close()

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		influxdbItems := make([]*dataobj.InfluxDBItem, count)
		for i := 0; i < count; i++ {
			influxdbItems[i] = items[i].(*dataobj.InfluxDBItem)
			stats.Counter.Set("points.out.influxdb", 1)
			logger.Debug("send to influxdb: ", influxdbItems[i])
		}

		//  同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(addr string, influxdbItems []*dataobj.InfluxDBItem, count int) {
			defer sema.Release()
			sendOk := false

			for i := 0; i < retry; i++ {
				err = influxClient.Send(influxdbItems)
				if err == nil {
					sendOk = true
					break
				}
				logger.Warningf("send influxdb fail: %v", err)
				time.Sleep(time.Millisecond * 10)
			}

			if !sendOk {
				stats.Counter.Set("points.out.influxdb.err", count)
				logger.Errorf("send %v to influxdb %s:%s fail: %v", influxdbItems, node, addr, err)
			} else {
				logger.Debugf("send to influxdb %s:%s ok", node, addr)
			}
		}(addr, influxdbItems, count)
	}
}
