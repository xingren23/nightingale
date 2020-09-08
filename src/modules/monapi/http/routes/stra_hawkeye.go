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
	"github.com/didi/nightingale/src/modules/monapi/config"
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
			logger.Errorf("get node err:%v %v", err, stra)
			continue
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

type metricsQueryForm struct {
	Nid   int64  `json:"nid"`
	Limit int    `json:"limit"`
	Query string `json:"query"`
}

type metricResp struct {
	Metric string `json:"metric"`
	Note   string `json:"note"`
}

func straMetricsPost(c *gin.Context) {
	var f metricsQueryForm
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
		metrics, err := getMetricsByTransfer(qEndpoints)
		errors.Dangerous(err)

		for _, metric := range metrics {
			if f.Query != "" && !strings.Contains(metric, f.Query) {
				continue
			}
			var note string
			if item, exists := mcache.MetricInfoCache.Get(metric); exists {
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
		metricInfoMap := mcache.MetricInfoCache.GetAll()
		for _, item := range metricInfoMap {
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

func getMetricsByTransfer(endpoints []string) ([]string, error) {
	addrs := address.GetHTTPAddresses("transfer")
	if len(addrs) == 0 {
		return nil, fmt.Errorf("empty transfer addr")
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

type tagKsQueryForm struct {
	Nid    int64  `json:"nid"`
	Metric string `json:"metric"`
}

func straTagKeysPost(c *gin.Context) {
	var f tagKsQueryForm
	errors.Dangerous(c.ShouldBind(&f))

	curNode, err := cmdb.GetCmdb().NodeGet("id", f.Nid)
	errors.Dangerous(err)
	// 补充服务树标签
	res := []string{config.FilterTagEnv, config.FilterTagHost, config.FilterTagNodePath}
	qEndpoints := make([]string, 0)
	if curNode.Leaf == 1 {
		leafIds, err := cmdb.GetCmdb().LeafIds(curNode)
		errors.Dangerous(err)

		endpoints, err := cmdb.GetCmdb().EndpointUnderLeafs(leafIds)
		errors.Dangerous(err)

		for _, e := range endpoints {
			qEndpoints = append(qEndpoints, e.Ident)
		}
	}

	tagKResps, err := getTagKeysByTransfer(f.Metric, qEndpoints)
	errors.Dangerous(err)

	for _, resp := range tagKResps {
		if resp.Metric == f.Metric {
			for _, k := range resp.TagKeys {
				if k == config.FilterTagEnv || k == "endpoint" || k == "ip" {
					continue
				}
				res = append(res, k)
			}
		}
	}

	renderData(c, res, err)
}

type TagKeysResp struct {
	Data []dataobj.TagKeysResp `json:"dat"`
	Err  string                `json:"err"`
}

func getTagKeysByTransfer(metric string, endpoints []string) ([]dataobj.TagKeysResp, error) {
	addrs := address.GetHTTPAddresses("transfer")
	if len(addrs) == 0 {
		return nil, fmt.Errorf("empty transfer addr")
	}

	var result TagKeysResp
	req := dataobj.EndpointMetricRecv{Endpoints: endpoints, Metrics: []string{metric}}
	perm := rand.Perm(len(addrs))
	var err error
	for i := range perm {
		url := fmt.Sprintf("http://%s%s", addrs[perm[i]], "/api/index/tagKeys")
		err = httplib.Post(url).JSONBodyQuiet(req).SetTimeout(time.Duration(500) * time.Millisecond).ToJSON(&result)
		if err == nil {
			break
		}
		logger.Warningf("get tag keys from transfer failed, error:%v, req:%+v", err, req)
	}

	if err != nil {
		return nil, fmt.Errorf("get tag keys from transfer failed, error:%v, req:%+v", err, req)
	}
	if result.Err != "" {
		return nil, fmt.Errorf(result.Err)
	}
	return result.Data, nil
}

type tagValsQueryForm struct {
	Nid      int64              `json:"nid"`
	Metric   string             `json:"metric"`
	Include  []*dataobj.TagPair `json:"include"`
	Exclude  []*dataobj.TagPair `json:"exclude"`
	QueryKey string             `json:"queryKey"`
	QueryVal string             `json:"queryVal"`
	Limit    int                `json:"limit"`
}

func straTagValsPost(c *gin.Context) {
	var f tagValsQueryForm
	errors.Dangerous(c.ShouldBind(&f))

	limit := 50
	if f.Limit > 0 {
		limit = f.Limit
	}

	if f.QueryKey == "" {
		errors.Dangerous("查询标签为空")
	}

	res := make([]string, 0)
	// 查询环境标签，写死
	if f.QueryKey == config.FilterTagEnv {
		envs := []string{"test", "stage", "prod"}
		renderData(c, likeSearch(envs, f.QueryVal, limit), nil)
		return
	}

	curNode, err := cmdb.GetCmdb().NodeGet("id", f.Nid)
	errors.Dangerous(err)

	leafNids, err := cmdb.GetCmdb().LeafIds(curNode)
	errors.Dangerous(err)

	qEndpoints := make([]string, 0)
	// 查询host标签
	if f.QueryKey == config.FilterTagHost {
		endpoints, _, err := cmdb.GetCmdb().EndpointUnderNodeGets(leafNids, f.QueryVal, "", "", f.Limit, offset(c, f.Limit,
			999))
		errors.Dangerous(err)

		for _, e := range endpoints {
			qEndpoints = append(qEndpoints, e.Ident)
		}
		renderData(c, likeSearch(qEndpoints, f.QueryVal, limit), nil)
		return
	}

	if curNode.Leaf == 1 {
		endpoints, _, err := cmdb.GetCmdb().EndpointUnderNodeGets(leafNids, "", "", "", 100, offset(c, f.Limit,
			999))
		errors.Dangerous(err)

		for _, e := range endpoints {
			qEndpoints = append(qEndpoints, e.Ident)
		}
	}
	// 排除nodePath标签
	for idx, include := range f.Include {
		if include.Key == config.FilterTagNodePath || include.Key == config.FilterTagHost {
			f.Include = append(f.Include[:idx], f.Include[idx+1:]...)
			break
		}
	}
	for idx, exclude := range f.Exclude {
		if exclude.Key == config.FilterTagNodePath || exclude.Key == config.FilterTagHost {
			f.Exclude = append(f.Exclude[:idx], f.Exclude[idx+1:]...)
			break
		}
	}

	req := dataobj.TagValsCludeRecv{
		Endpoints: qEndpoints,
		Metric:    f.Metric,
		Include:   f.Include,
		Exclude:   f.Exclude,
		QueryPair: []*dataobj.TagPair{{Key: f.QueryKey, Values: []string{f.QueryVal}}},
		Limit:     limit,
	}
	tagVResp, err := getTagValsByTransfer(req)
	errors.Dangerous(err)

	for _, tag := range tagVResp.Tagkvs {
		res = append(res, tag.Values...)
	}

	renderData(c, res, err)
}

type TagValsResp struct {
	Data *dataobj.TagValsXcludeResp `json:"dat"`
	Err  string                     `json:"err"`
}

func getTagValsByTransfer(req dataobj.TagValsCludeRecv) (*dataobj.TagValsXcludeResp, error) {
	addrs := address.GetHTTPAddresses("transfer")
	if len(addrs) == 0 {
		return nil, fmt.Errorf("empty transfer addr")
	}

	var result TagValsResp
	perm := rand.Perm(len(addrs))
	var err error
	for i := range perm {
		url := fmt.Sprintf("http://%s%s", addrs[perm[i]], "/api/index/tagVals")
		err = httplib.Post(url).JSONBodyQuiet(req).SetTimeout(time.Duration(500) * time.Millisecond).ToJSON(&result)
		if err == nil {
			break
		}
		logger.Warningf("get tag values from transfer failed, error:%v, req:%+v", err, req)
	}

	if err != nil {
		return nil, fmt.Errorf("get tag values from transfer failed, error:%v, req:%+v", err, req)
	}
	if result.Err != "" {
		return nil, fmt.Errorf(result.Err)
	}
	if result.Data == nil {
		return nil, fmt.Errorf("get tag values from transfer failed, error: data is nil")
	}
	return result.Data, nil
}

func likeSearch(items []string, query string, limit int) []string {
	res := make([]string, 0)
	var count int
	for _, item := range items {
		if query != "" && !strings.Contains(item, query) {
			continue
		}
		count++
		if count >= limit {
			break
		}
		res = append(res, item)
	}
	return res
}

func straEffectiveGet(c *gin.Context) {
	sid := queryInt64(c, "id", 0)
	stra, err := model.StraGet("id", sid)
	errors.Dangerous(err)

	node, err := scache.JudgeHashRing.GetNode(strconv.FormatInt(stra.Id, 10))
	errors.Dangerous(err)

	stras := scache.StraCache.GetByNode(node)
	if stras != nil && len(stras) > 0 {
		for _, item := range stras {
			if item.Id == stra.Id {
				stra.Endpoints = item.Endpoints
				stra.Tags = item.Tags
				judgeNode, err := scache.JudgeActiveNode.GetInstanceBy(node)
				errors.Dangerous(err)

				stra.JudgeInstance = judgeNode
				break
			}
		}
	}
	renderData(c, stra, nil)
}
