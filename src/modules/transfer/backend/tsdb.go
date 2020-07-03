package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/didi/nightingale/src/modules/transfer/calc"

	"github.com/didi/nightingale/src/toolkits/address"
	"github.com/toolkits/pkg/net/httplib"

	"github.com/toolkits/pkg/pool"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/toolkits/pools"
	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/toolkits/pkg/concurrent/semaphore"
	"github.com/toolkits/pkg/container/list"
	"github.com/toolkits/pkg/container/set"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/str"
)

type TSDBStorage struct {
	//config
	section TsdbSection

	// 服务节点的一致性哈希环 pk -> node
	TsdbNodeRing *ConsistentHashRing

	// 发送缓存队列 node -> queue_of_data
	TsdbQueues map[string]*list.SafeListLimited

	// 连接池 node_address -> connection_pool
	TsdbConnPools *pools.ConnPools
}

func (tsdb *TSDBStorage) Init() {

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

	RegisterStorage(tsdb.section.Name, tsdb)
}

func (tsdb *TSDBStorage) QueryData(inputs []dataobj.QueryData) []*dataobj.TsdbQueryResponse {
	workerNum := 100
	worker := make(chan struct{}, workerNum) // 控制 goroutine 并发数
	dataChan := make(chan *dataobj.TsdbQueryResponse, 20000)

	done := make(chan struct{}, 1)
	resp := make([]*dataobj.TsdbQueryResponse, 0)
	go func() {
		defer func() { done <- struct{}{} }()
		for d := range dataChan {
			resp = append(resp, d)
		}
	}()

	for _, input := range inputs {
		for _, endpoint := range input.Endpoints {
			for _, counter := range input.Counters {
				worker <- struct{}{}
				go tsdb.fetchDataSync(input.Start, input.End, input.ConsolFunc, endpoint, counter, input.Step, worker,
					dataChan)
			}
		}
	}

	// 等待所有 goroutine 执行完成
	for i := 0; i < workerNum; i++ {
		worker <- struct{}{}
	}
	close(dataChan)

	// 等待所有 dataChan 被消费完
	<-done

	return resp
}

// Push2TsdbSendQueue pushes data to a TSDB instance which depends on the consistent ring.
func (tsdb *TSDBStorage) Push2Queue(items []*dataobj.MetricValue) {
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

func (tsdb *TSDBStorage) Send2TsdbTask(Q *list.SafeListLimited, node, addr string, concurrent int) {
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
func (tsdb *TSDBStorage) convert2TsdbItem(d *dataobj.MetricValue) *dataobj.TsdbItem {
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

func (tsdb *TSDBStorage) QueryDataForUI(input dataobj.QueryDataForUI) []*dataobj.TsdbQueryResponse {
	workerNum := 100
	worker := make(chan struct{}, workerNum) // 控制 goroutine 并发数
	dataChan := make(chan *dataobj.TsdbQueryResponse, 20000)

	done := make(chan struct{}, 1)
	resp := make([]*dataobj.TsdbQueryResponse, 0)
	go func() {
		defer func() { done <- struct{}{} }()
		for d := range dataChan {
			resp = append(resp, d)
		}
	}()

	for _, endpoint := range input.Endpoints {
		if len(input.Tags) == 0 {
			counter, err := tsdb.GetCounter(input.Metric, "", nil)
			if err != nil {
				logger.Warningf("get counter error: %+v", err)
				continue
			}
			worker <- struct{}{}
			go tsdb.fetchDataSync(input.Start, input.End, input.ConsolFunc, endpoint, counter, input.Step, worker,
				dataChan)
		} else {
			for _, tag := range input.Tags {
				counter, err := tsdb.GetCounter(input.Metric, tag, nil)
				if err != nil {
					logger.Warningf("get counter error: %+v", err)
					continue
				}
				worker <- struct{}{}
				go tsdb.fetchDataSync(input.Start, input.End, input.ConsolFunc, endpoint, counter, input.Step, worker,
					dataChan)
			}
		}
	}

	// 等待所有 goroutine 执行完成
	for i := 0; i < workerNum; i++ {
		worker <- struct{}{}
	}

	close(dataChan)
	<-done

	//进行数据计算
	aggrDatas := make([]*dataobj.TsdbQueryResponse, 0)
	if input.AggrFunc != "" && len(resp) > 1 {
		aggrCounter := make(map[string][]*dataobj.TsdbQueryResponse)

		// 没有聚合 tag, 或者曲线没有其他 tags, 直接所有曲线进行计算
		if len(input.GroupKey) == 0 || getTags(resp[0].Counter) == "" {
			aggrData := &dataobj.TsdbQueryResponse{
				Start:  input.Start,
				End:    input.End,
				Values: calc.Compute(input.AggrFunc, resp),
			}
			aggrDatas = append(aggrDatas, aggrData)
		} else {
			for _, data := range resp {
				counterMap := make(map[string]string)

				tagsMap, err := dataobj.SplitTagsString(getTags(data.Counter))
				if err != nil {
					logger.Warningf("split tag string error: %+v", err)
					continue
				}
				tagsMap["endpoint"] = data.Endpoint

				// 校验 GroupKey 是否在 tags 中
				for _, key := range input.GroupKey {
					if value, exists := tagsMap[key]; exists {
						counterMap[key] = value
					}
				}

				counter := dataobj.SortedTags(counterMap)
				if _, exists := aggrCounter[counter]; exists {
					aggrCounter[counter] = append(aggrCounter[counter], data)
				} else {
					aggrCounter[counter] = []*dataobj.TsdbQueryResponse{data}
				}
			}

			// 有需要聚合的 tag 需要将 counter 带上
			for counter, datas := range aggrCounter {
				aggrData := &dataobj.TsdbQueryResponse{
					Start:   input.Start,
					End:     input.End,
					Counter: counter,
					Values:  calc.Compute(input.AggrFunc, datas),
				}
				aggrDatas = append(aggrDatas, aggrData)
			}
		}
		return aggrDatas
	}
	return resp
}

func getTags(counter string) (tags string) {
	idx := strings.IndexAny(counter, "/")
	if idx == -1 {
		return ""
	}
	return counter[idx+1:]
}

func (tsdb *TSDBStorage) GetCounter(metric, tag string, tagMap map[string]string) (counter string, err error) {
	if tagMap == nil {
		tagMap, err = dataobj.SplitTagsString(tag)
		if err != nil {
			logger.Warningf("split tag string error: %+v", err)
			return
		}
	}

	tagStr := dataobj.SortedTags(tagMap)
	counter = dataobj.PKWithTags(metric, tagStr)
	return
}

func (tsdb *TSDBStorage) fetchDataSync(start, end int64, consolFun, endpoint, counter string, step int, worker chan struct{}, dataChan chan *dataobj.TsdbQueryResponse) {
	defer func() {
		<-worker
	}()
	stats.Counter.Set("query.tsdb", 1)

	data, err := tsdb.fetchData(start, end, consolFun, endpoint, counter, step)
	if err != nil {
		logger.Warningf("fetch tsdb data error: %+v", err)
		stats.Counter.Set("query.data.err", 1)
		data.Endpoint = endpoint
		data.Counter = counter
		data.Step = step
	}
	dataChan <- data
}

func (tsdb *TSDBStorage) fetchData(start, end int64, consolFun, endpoint, counter string, step int) (*dataobj.TsdbQueryResponse, error) {
	var resp *dataobj.TsdbQueryResponse

	qparm := GenQParam(start, end, consolFun, endpoint, counter, step)
	resp, err := tsdb.QueryOne(qparm)
	if err != nil {
		return resp, err
	}

	resp.Start = start
	resp.End = end

	return resp, nil
}

func GenQParam(start, end int64, consolFunc, endpoint, counter string, step int) dataobj.TsdbQueryParam {
	return dataobj.TsdbQueryParam{
		Start:      start,
		End:        end,
		ConsolFunc: consolFunc,
		Endpoint:   endpoint,
		Counter:    counter,
		Step:       step,
	}
}

func (tsdb *TSDBStorage) QueryOne(para dataobj.TsdbQueryParam) (resp *dataobj.TsdbQueryResponse, err error) {
	start, end := para.Start, para.End
	resp = &dataobj.TsdbQueryResponse{}

	pk := dataobj.PKWithCounter(para.Endpoint, para.Counter)
	ps, err := tsdb.SelectPoolByPK(pk)
	if err != nil {
		return resp, err
	}

	count := len(ps)
	for _, i := range rand.Perm(count) {
		onePool := ps[i].Pool
		addr := ps[i].Addr

		conn, err := onePool.Fetch()
		if err != nil {
			logger.Errorf("fetch pool error: %+v", err)
			continue
		}

		rpcConn := conn.(pools.RpcClient)
		if rpcConn.Closed() {
			onePool.ForceClose(conn)

			err = errors.New("conn closed")
			logger.Error(err)
			continue
		}

		type ChResult struct {
			Err  error
			Resp *dataobj.TsdbQueryResponse
		}

		ch := make(chan *ChResult, 1)
		go func() {
			resp := &dataobj.TsdbQueryResponse{}
			err := rpcConn.Call("Tsdb.Query", para, resp)
			ch <- &ChResult{Err: err, Resp: resp}
		}()

		select {
		case <-time.After(time.Duration(tsdb.section.CallTimeout) * time.Millisecond):
			onePool.ForceClose(conn)
			logger.Errorf("%s, call timeout. proc: %s", addr, onePool.Proc())
			break
		case r := <-ch:
			if r.Err != nil {
				onePool.ForceClose(conn)
				logger.Errorf("%s, call failed, err %v. proc: %s", addr, r.Err, onePool.Proc())
				break
			} else {
				onePool.Release(conn)
				if len(r.Resp.Values) < 1 {
					r.Resp.Values = []*dataobj.RRDData{}
					return r.Resp, nil
				}

				fixed := make([]*dataobj.RRDData, 0)
				for _, v := range r.Resp.Values {
					if v == nil || !(v.Timestamp >= start && v.Timestamp <= end) {
						continue
					}
					fixed = append(fixed, v)
				}
				r.Resp.Values = fixed
			}
			return r.Resp, nil
		}

	}
	return resp, fmt.Errorf("get data error")

}

type Pool struct {
	Pool *pool.ConnPool
	Addr string
}

func (tsdb *TSDBStorage) SelectPoolByPK(pk string) ([]Pool, error) {
	node, err := tsdb.TsdbNodeRing.GetNode(pk)
	if err != nil {
		return []Pool{}, err
	}

	nodeAddrs, found := tsdb.section.ClusterList[node]
	if !found {
		return []Pool{}, errors.New("node not found")
	}

	var ps []Pool
	for _, addr := range nodeAddrs.Addrs {
		onePool, found := tsdb.TsdbConnPools.Get(addr)
		if !found {
			logger.Errorf("addr %s not found", addr)
			continue
		}
		ps = append(ps, Pool{Pool: onePool, Addr: addr})
	}

	if len(ps) < 1 {
		return ps, errors.New("addr not found")
	}

	return ps, nil
}

type Tagkv struct {
	TagK string   `json:"tagk"`
	TagV []string `json:"tagv"`
}

type SeriesReq struct {
	Endpoints []string `json:"endpoints"`
	Metric    string   `json:"metric"`
	Tagkv     []*Tagkv `json:"tagkv"`
}

type SeriesResp struct {
	Dat []Series `json:"dat"`
	Err string   `json:"err"`
}

type Series struct {
	Endpoints []string `json:"endpoints"`
	Metric    string   `json:"metric"`
	Tags      []string `json:"tags"`
	Step      int      `json:"step"`
	DsType    string   `json:"dstype"`
}

func (tsdb *TSDBStorage) QuerySeries(start, end int64, req []SeriesReq) ([]dataobj.QueryData, error) {
	var res SeriesResp
	var queryDatas []dataobj.QueryData

	if len(req) < 1 {
		return queryDatas, fmt.Errorf("req length < 1")
	}

	addrs := address.GetHTTPAddresses("index")

	if len(addrs) < 1 {
		return queryDatas, fmt.Errorf("index addr is nil")
	}

	i := rand.Intn(len(addrs))
	addr := fmt.Sprintf("http://%s/api/index/counter/fullmatch", addrs[i])

	resp, code, err := httplib.PostJSON(addr, time.Duration(tsdb.section.IndexTimeout)*time.Millisecond, req, nil)
	if err != nil {
		return queryDatas, err
	}

	if code != 200 {
		return nil, fmt.Errorf("index response status code != 200")
	}

	if err = json.Unmarshal(resp, &res); err != nil {
		logger.Error(string(resp))
		return queryDatas, err
	}

	for _, item := range res.Dat {
		counters := make([]string, 0)
		if len(item.Tags) == 0 {
			counters = append(counters, item.Metric)
		} else {
			for _, tag := range item.Tags {
				tagMap, err := dataobj.SplitTagsString(tag)
				if err != nil {
					logger.Warning(err, tag)
					continue
				}
				tagStr := dataobj.SortedTags(tagMap)
				counter := dataobj.PKWithTags(item.Metric, tagStr)
				counters = append(counters, counter)
			}
		}

		queryData := dataobj.QueryData{
			Start:      start,
			End:        end,
			Endpoints:  item.Endpoints,
			Counters:   counters,
			ConsolFunc: "AVERAGE",
			DsType:     item.DsType,
			Step:       item.Step,
		}
		queryDatas = append(queryDatas, queryData)
	}

	return queryDatas, err
}
