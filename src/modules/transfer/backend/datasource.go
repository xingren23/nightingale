package backend

import (
	"fmt"
	"time"

	"github.com/didi/nightingale/src/dataobj"
)

// send
const (
	DefaultSendTaskSleepInterval = time.Millisecond * 50 //默认睡眠间隔为50ms
	DefaultSendQueueMaxSize      = 102400                //10.24w
	MaxSendRetry                 = 10
)

var (
	MinStep int //最小上报周期,单位sec
)

type Datasource interface {
	PushEndpoint

	// query data for judge
	QueryData(inputs []dataobj.QueryData) []*dataobj.TsdbQueryResponse
	// query data for ui
	QueryDataForUI(input dataobj.QueryDataForUI) []*dataobj.TsdbQueryResponse

	// query metrics & tags
	QueryMetrics(recv dataobj.EndpointsRecv) *dataobj.MetricResp
	QueryTagPairs(recv dataobj.EndpointMetricRecv) []dataobj.IndexTagkvResp
	QueryIndexByClude(recv []dataobj.CludeRecv) []dataobj.XcludeResp
	QueryIndexByFullTags(recv []dataobj.IndexByFullTagsRecv) []dataobj.IndexByFullTagsResp

	// tsdb instance
	GetInstance(metric, endpoint string, tags map[string]string) []string
}

type PushEndpoint interface {
	// push data
	Push2Queue(items []*dataobj.MetricValue)
}

var registryStorages map[string]Datasource
var registryPushEndpoints map[string]PushEndpoint

func init() {
	registryStorages = make(map[string]Datasource)
	registryPushEndpoints = make(map[string]PushEndpoint)
}

// get default backend storage
func GetStorageFor(pluginId string) (Datasource, error) {
	if pluginId == "" {
		pluginId = defaultStorage
	}
	if storage, exists := registryStorages[pluginId]; exists {
		return storage, nil
	}
	return nil, fmt.Errorf("Could not find storage for plugin: %s ", pluginId)
}

// get all push endpoints
func GetPushEndpoints() ([]PushEndpoint, error) {
	if len(registryPushEndpoints) > 0 {
		items := make([]PushEndpoint, 0, len(registryPushEndpoints))
		for _, value := range registryPushEndpoints {
			items = append(items, value)
		}
		return items, nil
	}
	return nil, fmt.Errorf("Could not find pushendpoint ")
}

func RegisterStorage(pluginId string, storage Datasource) {

	registryStorages[pluginId] = storage
	registryPushEndpoints[pluginId] = storage
}

func RegisterPushEndpoint(pluginId string, push PushEndpoint) {
	registryPushEndpoints[pluginId] = push
}
