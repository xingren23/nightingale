package backend

import (
	"github.com/toolkits/pkg/container/list"
	"github.com/toolkits/pkg/container/set"
	"github.com/toolkits/pkg/str"

	"github.com/didi/nightingale/src/modules/transfer/cache"
	"github.com/didi/nightingale/src/toolkits/pools"
	"github.com/didi/nightingale/src/toolkits/report"
	"github.com/didi/nightingale/src/toolkits/stats"
)

type InfluxdbSection struct {
	Enabled     bool                    `yaml:"enabled"`
	Batch       int                     `yaml:"batch"`
	MaxRetry    int                     `yaml:"maxRetry"`
	WorkerNum   int                     `yaml:"workerNum"`
	Timeout     int                     `yaml:"timeout"`
	Database    string                  `yaml:"database"`
	Username    string                  `yaml:"username"`
	Password    string                  `yaml:"password"`
	Precision   string                  `yaml:"precision"`
	Replicas    int                     `yaml:"replicas"`
	Cluster     map[string]string       `yaml:"cluster"`
	ClusterList map[string]*ClusterNode `json:"clusterList"`
}

type TsdbSection struct {
	Enabled     bool                    `yaml:enabled`
	Replicas    int                     `yaml:"replicas"`
	Cluster     map[string]string       `yaml:"cluster"`
	ClusterList map[string]*ClusterNode `json:"clusterList"`
}

type BackendSection struct {
	Enabled      bool            `yaml:"enabled"`
	Batch        int             `yaml:"batch"`
	ConnTimeout  int             `yaml:"connTimeout"`
	CallTimeout  int             `yaml:"callTimeout"`
	WorkerNum    int             `yaml:"workerNum"`
	MaxConns     int             `yaml:"maxConns"`
	MaxIdle      int             `yaml:"maxIdle"`
	IndexTimeout int             `yaml:"indexTimeout"`
	StraPath     string          `yaml:"straPath"`
	HbsMod       string          `yaml:"hbsMod"`
	Tsdb         TsdbSection     `yaml:tsdb`
	Influxdb     InfluxdbSection `yaml:"influxdb"`
}

const DefaultSendQueueMaxSize = 102400 //10.24w

type ClusterNode struct {
	Addrs []string `json:"addrs"`
}

var (
	Config BackendSection
	// 服务节点的一致性哈希环 pk -> node
	TsdbNodeRing   *ConsistentHashRing
	InfluxNodeRing *ConsistentHashRing

	// 发送缓存队列 node -> queue_of_data
	JudgeQueues   = cache.SafeJudgeQueue{}
	TsdbQueues    = make(map[string]*list.SafeListLimited)
	InfluxdbQueue = make(map[string]*list.SafeListLimited)

	// 连接池 node_address -> connection_pool
	TsdbConnPools  *pools.ConnPools
	JudgeConnPools *pools.ConnPools
)

func Init(cfg BackendSection) {
	Config = cfg

	initHashRing()
	initConnPools()
	initSendQueues()

	startSendTasks()
}

func initHashRing() {
	TsdbNodeRing = NewConsistentHashRing(int32(Config.Tsdb.Replicas), str.KeysOfMap(Config.Tsdb.Cluster))

	InfluxNodeRing = NewConsistentHashRing(int32(Config.Influxdb.Replicas), str.KeysOfMap(Config.Influxdb.Cluster))
}

func initConnPools() {
	tsdbInstances := set.NewSafeSet()
	for _, item := range Config.Tsdb.ClusterList {
		for _, addr := range item.Addrs {
			tsdbInstances.Add(addr)
		}
	}
	TsdbConnPools = pools.NewConnPools(
		Config.MaxConns, Config.MaxIdle, Config.ConnTimeout, Config.CallTimeout, tsdbInstances.ToSlice(),
	)

	JudgeConnPools = pools.NewConnPools(
		Config.MaxConns, Config.MaxIdle, Config.ConnTimeout, Config.CallTimeout, GetJudges(),
	)
}

func initSendQueues() {
	JudgeQueues = cache.NewJudgeQueue()
	judges := GetJudges()
	for _, judge := range judges {
		JudgeQueues.Set(judge, list.NewSafeListLimited(DefaultSendQueueMaxSize))
	}

	for node, item := range Config.Tsdb.ClusterList {
		for _, addr := range item.Addrs {
			TsdbQueues[node+addr] = list.NewSafeListLimited(DefaultSendQueueMaxSize)
		}
	}

	for node, item := range Config.Influxdb.ClusterList {
		for _, addr := range item.Addrs {
			InfluxdbQueue[node+addr] = list.NewSafeListLimited(DefaultSendQueueMaxSize)
		}
	}
}

func GetJudges() []string {
	var judgeInstances []string
	instances, err := report.GetAlive("judge", Config.HbsMod)
	if err != nil {
		stats.Counter.Set("judge.get.err", 1)
		return judgeInstances
	}
	for _, instance := range instances {
		judgeInstance := instance.Identity + ":" + instance.RPCPort
		judgeInstances = append(judgeInstances, judgeInstance)
	}
	return judgeInstances
}
