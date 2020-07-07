package backend

import (
	"strings"
	"time"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/toolkits/pools"
	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/toolkits/pkg/concurrent/semaphore"
	"github.com/toolkits/pkg/container/list"
	"github.com/toolkits/pkg/container/set"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/str"
)

type TsdbStorage struct {
	//config
	section TsdbSection

	// 服务节点的一致性哈希环 pk -> node
	TsdbNodeRing *ConsistentHashRing

	// 发送缓存队列 node -> queue_of_data
	TsdbQueues map[string]*list.SafeListLimited

	// 连接池 node_address -> connection_pool
	TsdbConnPools *pools.ConnPools
}

func (tsdb *TsdbStorage) Init() {

	// init hash ring
	tsdb.TsdbNodeRing = NewConsistentHashRing(int32(tsdb.section.Replicas),
		str.KeysOfMap(tsdb.section.Cluster))

	// init connPool
	tsdbInstances := set.NewSafeSet()
	for _, item := range tsdb.section.ClusterList {
		for _, addr := range item.Addrs {
			tsdbInstances.Add(addr)
		}
	}
	tsdb.TsdbConnPools = pools.NewConnPools(
		tsdb.section.MaxConns, tsdb.section.MaxIdle, tsdb.section.ConnTimeout, tsdb.section.CallTimeout,
		tsdbInstances.ToSlice(),
	)

	// init queues
	tsdb.TsdbQueues = make(map[string]*list.SafeListLimited)
	for node, item := range tsdb.section.ClusterList {
		for _, addr := range item.Addrs {
			tsdb.TsdbQueues[node+addr] = list.NewSafeListLimited(DefaultSendQueueMaxSize)
		}
	}

	// start task
	tsdbConcurrent := tsdb.section.WorkerNum
	if tsdbConcurrent < 1 {
		tsdbConcurrent = 1
	}
	for node, item := range tsdb.section.ClusterList {
		for _, addr := range item.Addrs {
			queue := tsdb.TsdbQueues[node+addr]
			go tsdb.Send2TsdbTask(queue, node, addr, tsdbConcurrent)
		}
	}

	go GetIndexLoop()

	RegisterStorage(tsdb.section.Name, tsdb)
}

// Push2TsdbSendQueue pushes data to a TSDB instance which depends on the consistent ring.
func (tsdb *TsdbStorage) Push2Queue(items []*dataobj.MetricValue) {
	errCnt := 0
	for _, item := range items {
		tsdbItem := tsdb.convert2TsdbItem(item)
		stats.Counter.Set("tsdb.queue.push", 1)

		node, err := tsdb.TsdbNodeRing.GetNode(item.PK())
		if err != nil {
			logger.Warningf("get tsdb node error: %v", err)
			continue
		}

		cnode := tsdb.section.ClusterList[node]
		for _, addr := range cnode.Addrs {
			Q := tsdb.TsdbQueues[node+addr]
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

func (tsdb *TsdbStorage) Send2TsdbTask(Q *list.SafeListLimited, node, addr string, concurrent int) {
	batch := tsdb.section.Batch // 一次发送,最多batch条数据
	Q = tsdb.TsdbQueues[node+addr]

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
				err = tsdb.TsdbConnPools.Call(addr, "Tsdb.Send", tsdbItems, resp)
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

// 打到 Tsdb 的数据,要根据 rrdtool 的特定 来限制 step、counterType、timestamp
func (tsdb *TsdbStorage) convert2TsdbItem(d *dataobj.MetricValue) *dataobj.TsdbItem {
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

	return item
}

func getTags(counter string) (tags string) {
	idx := strings.IndexAny(counter, "/")
	if idx == -1 {
		return ""
	}
	return counter[idx+1:]
}
