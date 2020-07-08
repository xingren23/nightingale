package scache

import (
	"fmt"
	"github.com/didi/nightingale/src/modules/monapi/dataobj"
	"github.com/didi/nightingale/src/modules/monapi/ecache"
	"github.com/toolkits/pkg/container/set"
	"strconv"
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/didi/nightingale/src/model"
)

var JudgeHashRing *ConsistentHashRing
var JudgeActiveNode = NewNodeMap()

func Init() {
	// 初始化默认参数
	StraCache = NewStraCache()
	CollectCache = NewCollectCache()
	JudgeHashRing = NewConsistentHashRing(500, []string{})

	go SyncStras()
	go SyncCollects()
}

func SyncStras() {
	t1 := time.NewTicker(time.Duration(10) * time.Second)

	syncStras()
	logger.Info("[cron] sync stras start...")
	for {
		<-t1.C
		syncStras()
	}
}

func syncStras() {
	stras, err := model.EffectiveStrasList()
	if err != nil {
		logger.Error("sync stras err:", err)
		return
	}
	strasMap := make(map[string][]*model.Stra)
	for _, stra := range stras {
		//增加叶子节点nid
		//stra.LeafNids, err = GetLeafNids(stra.Nid, stra.ExclNid)
		//if err != nil {
		//	logger.Warningf("get LeafNids err:%v %v", err, stra)
		//	continue
		//}
		//
		//endpoints, err := model.EndpointUnderLeafs(stra.LeafNids)
		//if err != nil {
		//	logger.Warningf("get endpoints err:%v %v", err, stra)
		//	continue
		//}

		endpoints, err := GetEndpointsByStra(stra)
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, stra)
			continue
		}

		for _, e := range endpoints {
			stra.Endpoints = append(stra.Endpoints, e.Ident)
		}

		node, err := JudgeHashRing.GetNode(strconv.FormatInt(stra.Id, 10))
		if err != nil {
			logger.Warningf("get node err:%v %v", err, stra)
		}

		if _, exists := strasMap[node]; exists {
			strasMap[node] = append(strasMap[node], stra)
		} else {
			strasMap[node] = []*model.Stra{stra}
		}
	}

	StraCache.SetAll(strasMap)
}

func SyncCollects() {
	t1 := time.NewTicker(time.Duration(10) * time.Second)

	syncCollects()
	logger.Info("[cron] sync collects start...")
	for {
		<-t1.C
		syncCollects()
	}
}

func syncCollects() {
	collectMap := make(map[string]*model.Collect)

	ports, err := model.GetPortCollects()
	if err != nil {
		logger.Warningf("get port collects err:%v", err)
	}

	for _, p := range ports {
		//leafNids, err := GetLeafNids(p.Nid, []int64{})
		//if err != nil {
		//	logger.Warningf("get LeafNids err:%v %v", err, p)
		//	continue
		//}
		//
		//endpoints, err := model.EndpointUnderLeafs(leafNids)
		//if err != nil {
		//	logger.Warningf("get endpoints err:%v %v", err, p)
		//	continue
		//}

		endpoints, err := GetEndpointsByNid(p.Nid, "HOST")
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, p.Nid)
			continue
		}

		for _, endpoint := range endpoints {
			name := endpoint.Ident
			c, exists := collectMap[name]
			if !exists {
				c = model.NewCollect()
			}
			c.Ports[p.Port] = p

			collectMap[name] = c
		}
	}

	procs, err := model.GetProcCollects()
	if err != nil {
		logger.Warningf("get port collects err:%v", err)
	}

	for _, p := range procs {
		//leafNids, err := GetLeafNids(p.Nid, []int64{})
		//if err != nil {
		//	logger.Warningf("get LeafNids err:%v %v", err, p)
		//	continue
		//}
		//
		//endpoints, err := model.EndpointUnderLeafs(leafNids)
		//if err != nil {
		//	logger.Warningf("get endpoints err:%v %v", err, p)
		//	continue
		//}

		endpoints, err := GetEndpointsByNid(p.Nid, "HOST")
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, p.Nid)
			continue
		}

		for _, endpoint := range endpoints {
			name := endpoint.Ident
			c, exists := collectMap[name]
			if !exists {
				c = model.NewCollect()
			}
			c.Procs[p.Target] = p
			collectMap[name] = c
		}
	}

	logConfigs, err := model.GetLogCollects()
	if err != nil {
		logger.Warningf("get log collects err:%v", err)
	}

	for _, l := range logConfigs {
		l.Decode()
		//leafNids, err := GetLeafNids(l.Nid, []int64{})
		//if err != nil {
		//	logger.Warningf("get LeafNids err:%v %v", err, l)
		//	continue
		//}
		//
		//Endpoints, err := model.EndpointUnderLeafs(leafNids)
		//if err != nil {
		//	logger.Warningf("get endpoints err:%v %v", err, l)
		//	continue
		//}

		Endpoints, err := GetEndpointsByNid(l.Nid, "HOST")
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, l.Nid)
			continue
		}

		for _, endpoint := range Endpoints {
			name := endpoint.Ident
			c, exists := collectMap[name]
			if !exists {
				c = model.NewCollect()
			}
			c.Logs[l.Name] = l
			collectMap[name] = c
		}
	}

	pluginConfigs, err := model.GetPluginCollects()
	if err != nil {
		logger.Warningf("get log collects err:%v", err)
	}

	for _, p := range pluginConfigs {
		//leafNids, err := GetLeafNids(p.Nid, []int64{})
		//if err != nil {
		//	logger.Warningf("get LeafNids err:%v %v", err, p)
		//	continue
		//}
		//
		//Endpoints, err := model.EndpointUnderLeafs(leafNids)
		//if err != nil {
		//	logger.Warningf("get endpoints err:%v %v", err, p)
		//	continue
		//}

		Endpoints, err := GetEndpointsByNid(p.Nid, "HOST")
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, p.Nid)
			continue
		}

		for _, endpoint := range Endpoints {
			name := endpoint.Ident
			c, exists := collectMap[name]
			if !exists {
				c = model.NewCollect()
			}

			key := fmt.Sprintf("%s-%d", p.Name, p.Nid)
			c.Plugins[key] = p
			collectMap[name] = c
		}
	}

	CollectCache.SetAll(collectMap)
}

func GetLeafNids(nid int64, exclNid []int64) ([]int64, error) {
	leafIds := []int64{}
	idsMap := make(map[int64]bool)
	node, err := model.GetNodeById(nid)
	if err != nil {
		return leafIds, err
	}

	if node == nil {
		return nil, fmt.Errorf("no such node[%d]", nid)
	}

	ids, err := node.LeafIds()
	if err != nil {
		return leafIds, err
	}
	//排除节点为空，直接将所有叶子节点返回
	if len(exclNid) == 0 {
		return ids, nil
	}

	for _, id := range ids {
		idsMap[id] = true
	}
	for _, id := range exclNid {
		delete(idsMap, id)
	}

	for id := range idsMap {
		leafIds = append(leafIds, id)
	}
	return leafIds, err
}

func removeDuplicateElement(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	temp := map[string]struct{}{}
	for _, item := range addrs {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func GetEndpointsByStra(stra *model.Stra) ([]model.Endpoint, error) {
	if len(stra.Exprs) == 0 {
		return nil, fmt.Errorf("stra is nil or stra.Exprs size is zero")
	}

	//获取MonitorItem的类型
	item, exists := ecache.MonitorItemCache.Get(stra.Exprs[0].Metric)
	if !exists {
		return nil, fmt.Errorf("MonitorItem is not exists: metric:%v", stra.Exprs[0].Metric)
	}

	nodePath, exists := ecache.SrvTreeCache.Get(stra.Nid)
	if !exists {
		return nil, fmt.Errorf("nodePath is not exists: srvTreeId:%v", stra.Nid)
	}

	tagEndpoints, exists := ecache.SrvTagEndpointCache.GetByKey(nodePath, item.EndpointType)
	if !exists {
		return nil, fmt.Errorf("endpoints is not exists: nodePath:%v, endpointType:%v", nodePath, item.EndpointType)
	}

	endpointSets := filterEnvs(tagEndpoints, stra, item.EndpointType)
	endpointSets = filterNodeIds(endpointSets, stra, item.EndpointType)
	endpointSets = filterNodePath(endpointSets, stra, item.EndpointType)
	endpointSets = filterHost(endpointSets, stra)

	endpointList := make([]model.Endpoint, 0)
	for _, endpoint := range endpointSets.ToSlice() {
		endpointModel, exists := ecache.EndpointCache.Get(endpoint)
		if exists {
			endpointList = append(endpointList, *endpointModel)
		}
	}

	return endpointList, nil
}

func GetEndpointsByNid(nid int64, endpointType string) ([]model.Endpoint, error) {
	nodePath, exists := ecache.SrvTreeCache.Get(nid)
	if !exists {
		return nil, fmt.Errorf("GetEndpointsByNid nodePath is not exists: srvTreeId:%v", nid)
	}

	tagEndpoints, exists := ecache.SrvTagEndpointCache.GetByKey(nodePath, endpointType)
	if !exists {
		return nil, fmt.Errorf("GetEndpointsByNid endpoints is not exists: nodePath:%v, endpointType:%v", nodePath, endpointType)
	}

	endpointList := make([]model.Endpoint, 0)
	for _, tagEndpoint := range tagEndpoints {
		endpoint, exists := ecache.EndpointCache.Get(tagEndpoint.Endpoint)
		if exists {
			endpointList = append(endpointList, *endpoint)
		}
	}

	return endpointList, nil
}

func filterEnvs(tagEndpoints []*dataobj.TagEndpoint, stra *model.Stra, endpointType string) *set.StringSet {
	isContain, envCodes := analysisTag(stra, "env")

	endpointSets := set.NewStringSet()
	for _, tagEndpoint := range tagEndpoints {
		endpointSets.Add(tagEndpoint.Endpoint)
	}

	if len(envCodes.M) == 0 {
		return endpointSets
	}

	hosts := set.NewStringSet()
	for _, tagEndpoint := range tagEndpoints {
		for _, envCode := range envCodes.ToSlice() {
			if tagEndpoint.EnvCode == envCode {
				hosts.Add(tagEndpoint.Endpoint)
			}
		}
	}

	return buildEndpointSet(endpointSets, hosts, isContain)
}

func filterNodeIds(endpointSets *set.StringSet, stra *model.Stra, endpointType string) *set.StringSet {
	nids := stra.ExclNid
	if nids == nil || len(nids) == 0 {
		return endpointSets
	}

	hosts := set.NewStringSet()
	for _, nid := range nids {
		expression, exists := ecache.SrvTreeCache.Get(nid)
		if exists {
			tagEndpoints, exists := ecache.SrvTagEndpointCache.GetByKey(expression, endpointType)
			if exists {
				for _, tagEndpoint := range tagEndpoints {
					hosts.Add(tagEndpoint.Endpoint)
				}
			}
		}
	}

	return buildEndpointSet(endpointSets, hosts, false)
}

func buildEndpointSet(endpointSets *set.StringSet, hostSet *set.StringSet, isContain bool) *set.StringSet {
	endpoints := endpointSets.ToSlice()
	if isContain {
		for _, endpoint := range endpoints {
			if !hostSet.Exists(endpoint) {
				endpointSets.Delete(endpoint)
			}
		}
	} else {
		for _, endpoint := range endpoints {
			if hostSet.Exists(endpoint) {
				endpointSets.Delete(endpoint)
			}
		}
	}

	return endpointSets
}

func filterNodePath(endpointSets *set.StringSet, stra *model.Stra, endpointType string) *set.StringSet {
	isContain, nodePaths := analysisTag(stra, "nodePath")
	if len(nodePaths.M) == 0 {
		return endpointSets
	}

	hosts := set.NewStringSet()
	for _, nodePath := range nodePaths.ToSlice() {
		tagEndpoints, exists := ecache.SrvTagEndpointCache.GetByKey(nodePath, endpointType)
		if exists {
			for _, tagEndpoint := range tagEndpoints {
				hosts.Add(tagEndpoint.Endpoint)
			}

		}
	}

	return buildEndpointSet(endpointSets, hosts, isContain)
}

func filterHost(endpointSets *set.StringSet, stra *model.Stra) *set.StringSet {
	isContain, hosts := analysisTag(stra, "host")
	if len(hosts.M) == 0 {
		return endpointSets
	}

	return buildEndpointSet(endpointSets, hosts, isContain)
}

func analysisTag(stra *model.Stra, key string) (bool, *set.StringSet) {
	isContain := false
	tagValues := set.NewStringSet()
	for _, tag := range stra.Tags {
		if tag.Tkey == key {
			for _, value := range tag.Tval {
				tagValues.Add(value)
			}

			if tag.Topt == "=" {
				isContain = true
			}
		}
	}
	return isContain, tagValues
}
