package routes

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/cmdb"
	"github.com/didi/nightingale/src/modules/monapi/mcache"
	"github.com/didi/nightingale/src/modules/monapi/scache"
	"github.com/didi/nightingale/src/toolkits/address"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/net/httplib"
)

func strasHawkeyeGet(c *gin.Context) {
	name := queryStr(c, "name", "")
	priority := queryInt(c, "priority", 4)
	nid := queryInt64(c, "nid", 0)
	list, err := model.StrasList(name, priority, nid)
	straEndpointsMap := make(map[int64][]string)

	for _, stra := range list {
		node, err := scache.JudgeHashRing.GetNode(strconv.FormatInt(stra.Id, 10))
		if err != nil {
			logger.Warningf("get node err:%v %v", err, stra)
		}

		stras := scache.StraCache.GetByNode(node)
		if stras != nil && len(stras) > 0 {
			for _, straByCache := range stras {
				straEndpointsMap[straByCache.Id] = straByCache.Endpoints
			}
		}
	}

	for _, stra := range list {
		stra.Endpoints = straEndpointsMap[stra.Id]
	}
	renderData(c, list, err)
}

type metricsForm struct {
	Nid   int64  `json:"nid"`
	Limit int    `json:"limit"`
	Query string `json:"query"`
}

type metricResp struct {
	Metric string `json:"metric"`
	Note   string `json:"note"`
}

func straMetricsPost(c *gin.Context) {
	var f metricsForm
	errors.Dangerous(c.ShouldBind(&f))
	limit := 500
	if f.Limit > 0 {
		limit = f.Limit
	}
	curNode, err := cmdb.GetCmdb().NodeGet("id", f.Nid)
	errors.Dangerous(err)

	res := make([]metricResp, 0)
	idx := 0
	if curNode.Leaf == 1 {
		leafIds, err := cmdb.GetCmdb().LeafIds(curNode)
		errors.Dangerous(err)

		endpoints, err := cmdb.GetCmdb().EndpointUnderLeafs(leafIds)
		errors.Dangerous(err)

		qEndpoints := make([]string, 0)
		for _, e := range endpoints {
			qEndpoints = append(qEndpoints, e.Ident)
		}
		metrics, err := getMetricsByTranfer(qEndpoints)
		errors.Dangerous(err)

		for _, metric := range metrics {
			if f.Query != "" && !strings.Contains(metric, f.Query) {
				continue
			}
			var note string
			if item, exists := mcache.MonitorItemCache.Get(metric); exists {
				note = item.Description
			}
			m := metricResp{Metric: metric, Note: note}
			res = append(res, m)
			idx++
			if idx >= limit {
				break
			}
		}
	} else {
		monitorItemMap := mcache.MonitorItemCache.GetAll()
		for _, item := range monitorItemMap {
			if f.Query != "" && !strings.Contains(item.Metric, f.Query) {
				continue
			}
			m := metricResp{Metric: item.Metric, Note: item.Description}
			res = append(res, m)
			idx++
			if idx >= limit {
				break
			}
		}
	}

	renderData(c, res, err)
}

type MetricsResp struct {
	Data dataobj.MetricResp `json:"dat"`
	Err  string             `json:"err"`
}

func getMetricsByTranfer(endpoints []string) ([]string, error) {
	addrs := address.GetHTTPAddresses("transfer")
	if len(addrs) == 0 {
		return nil, fmt.Errorf("empty index addr")
	}

	var result MetricsResp
	req := dataobj.EndpointsRecv{Endpoints: endpoints}
	perm := rand.Perm(len(addrs))
	var err error
	for i := range perm {
		url := fmt.Sprintf("http://%s%s", addrs[perm[i]], "/api/index/metrics")
		err = httplib.Post(url).JSONBodyQuiet(req).SetTimeout(time.Duration(500) * time.Millisecond).ToJSON(&result)
		if err == nil {
			break
		}
		logger.Warningf("get metrics from transfer failed, error:%v, req:%+v", err, req)
	}

	if err != nil {
		return nil, fmt.Errorf("get metrics from transfer failed, error:%v, req:%+v", err, req)
	}
	if result.Err != "" {
		return nil, fmt.Errorf(result.Err)
	}
	return result.Data.Metrics, nil
}
