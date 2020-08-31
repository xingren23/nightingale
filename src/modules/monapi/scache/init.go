package scache

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/cmdb"
	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/didi/nightingale/src/modules/monapi/mcache"

	"github.com/toolkits/pkg/container/set"
	"github.com/toolkits/pkg/logger"
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
		//获取策略 endpoint type
		item, exists := mcache.MetricInfoCache.Get(stra.Exprs[0].Metric)
		if !exists {
			logger.Errorf("stra %s metric %s is not in metricInfo cache", stra.Name, stra.Exprs[0].Metric)
			continue
		}
		endpointType := buildEndpointType(item)

		// 环境标签
		envE, envN := analysisTag(stra, config.FilterTagEnv)
		hostE, hostN := analysisTag(stra, config.FilterTagHost)
		nodePathE, nodePathN := analysisTag(stra, config.FilterTagNodePath)

		//增加叶子节点nid(排除子节点)
		stra.LeafNids, err = GetLeafNids(stra.Nid, stra.ExclNid, nodePathE.ToSlice(), nodePathN.ToSlice())
		if err != nil {
			logger.Warningf("get LeafNids err:%v %v", err, stra)
			continue
		}

		endpoints, err := cmdb.GetCmdb().EndpointUnderLeafs(stra.LeafNids)
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, stra)
			continue
		}

		// 根据指标元数据类型加载 endpoint
		for _, e := range endpoints {
			// host filter
			if len(hostE.M) != 0 && !hostE.Exists(e.Ident) {
				continue
			}
			if len(hostN.M) != 0 && hostN.Exists(e.Ident) {
				continue
			}

			tags, err := dataobj.SplitTagsString(e.Tags)
			if err != nil {
				logger.Errorf("split endpoint %s tags %s error, %s", e.Ident, e.Tags, err)
				continue
			}
			// env filter
			if envTag, ok := tags["env"]; ok {
				if len(envE.M) != 0 && !envE.Exists(envTag) {
					continue
				}
				if len(envN.M) != 0 && envN.Exists(envTag) {
					continue
				}
			}

			//endpoint type filter
			if typeTag, ok := tags["type"]; ok {
				if typeTag == endpointType {
					stra.Endpoints = append(stra.Endpoints, e.Ident)
				}
			}
		}

		// drop filter tag
		if len(stra.Tags) > 0 {
			tagArrs := make([]model.Tag, 0)
			for _, tag := range stra.Tags {
				if tag.Tkey == config.FilterTagEnv || tag.Tkey == config.FilterTagHost || tag.Tkey == config.FilterTagNodePath {
					continue
				}
				tagArrs = append(tagArrs, tag)
			}
			stra.Tags = tagArrs
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
		leafNids, err := GetLeafNids(p.Nid, []int64{}, []string{}, []string{})
		if err != nil {
			logger.Warningf("get LeafNids err:%v %v", err, p)
			continue
		}

		endpoints, err := cmdb.GetCmdb().EndpointUnderLeafs(leafNids)
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, p)
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
		leafNids, err := GetLeafNids(p.Nid, []int64{}, []string{}, []string{})
		if err != nil {
			logger.Warningf("get LeafNids err:%v %v", err, p)
			continue
		}

		endpoints, err := cmdb.GetCmdb().EndpointUnderLeafs(leafNids)
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, p)
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
		leafNids, err := GetLeafNids(l.Nid, []int64{}, []string{}, []string{})
		if err != nil {
			logger.Warningf("get LeafNids err:%v %v", err, l)
			continue
		}

		Endpoints, err := cmdb.GetCmdb().EndpointUnderLeafs(leafNids)
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, l)
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
		leafNids, err := GetLeafNids(p.Nid, []int64{}, []string{}, []string{})
		if err != nil {
			logger.Warningf("get LeafNids err:%v %v", err, p)
			continue
		}

		Endpoints, err := cmdb.GetCmdb().EndpointUnderLeafs(leafNids)
		if err != nil {
			logger.Warningf("get endpoints err:%v %v", err, p)
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

// 支持排除节点串
func GetLeafNids(nid int64, exclNid []int64, includeNodes []string, excludeNodes []string) ([]int64, error) {
	leafIds := []int64{}
	idsMap := make(map[int64]bool)
	// 当前节点的所有子节点
	node, err := cmdb.GetCmdb().NodeGet("id", nid)
	if err != nil {
		return leafIds, err
	}

	if node == nil {
		return nil, fmt.Errorf("no such node[%d]", nid)
	}

	ids, err := cmdb.GetCmdb().LeafIds(node)
	if err != nil {
		return leafIds, err
	}
	for _, id := range ids {
		idsMap[id] = false
	}
	//排除节点为空，直接将所有叶子节点返回
	if len(exclNid) != 0 {
		exclLeafIds, err := GetExclLeafIds(exclNid)
		if err != nil {
			return leafIds, err
		}

		for _, id := range exclLeafIds {
			delete(idsMap, id)
		}
	}
	// includeNodes覆盖的所有节点
	if len(includeNodes) != 0 {
		for _, iPath := range includeNodes {
			pathNodes, err := cmdb.GetCmdb().NodeQueryPath(iPath, 9999)
			if err != nil {
				return nil, fmt.Errorf("query nodes by path error [%s]", iPath)
			}
			for _, pNode := range pathNodes {
				// 取和nid子节点交集
				if _, exists := idsMap[pNode.Id]; exists {
					idsMap[pNode.Id] = true
				}
			}
		}

		for id, exists := range idsMap {
			if exists == false {
				delete(idsMap, id)
			}
		}
	}

	if len(excludeNodes) != 0 {
		for _, ePath := range excludeNodes {
			pathNodes, err := cmdb.GetCmdb().NodeQueryPath(ePath, 9999)
			if err != nil {
				return nil, fmt.Errorf("query nodes by path error [%s]", ePath)
			}
			for _, pNode := range pathNodes {
				// 取和nid子节点交集
				if _, exists := idsMap[pNode.Id]; exists {
					delete(idsMap, pNode.Id)
				}
			}
		}
	}

	for id := range idsMap {
		leafIds = append(leafIds, id)
	}

	return leafIds, err
}

// GetExclLeafIds 获取排除节点下的叶子节点
func GetExclLeafIds(exclNid []int64) (leafIds []int64, err error) {
	for _, nid := range exclNid {
		node, err := cmdb.GetCmdb().NodeGet("id", nid)
		if err != nil {
			return leafIds, err
		}

		if node == nil {
			logger.Warningf("no such node[%d]", nid)
			continue
		}

		ids, err := cmdb.GetCmdb().LeafIds(node)
		if err != nil {
			return leafIds, err
		}
		leafIds = append(leafIds, ids...)
	}
	return leafIds, nil
}

func analysisTag(stra *model.Stra, key string) (equals *set.StringSet, notEquals *set.StringSet) {
	equals = set.NewStringSet()
	notEquals = set.NewStringSet()
	for _, tag := range stra.Tags {
		if tag.Tkey == key {
			if tag.Topt == "=" {
				for _, value := range tag.Tval {
					equals.Add(value)
				}
			} else if tag.Topt == "!=" {
				for _, value := range tag.Tval {
					notEquals.Add(value)
				}
			}
		}
	}
	return equals, notEquals
}

// TODO : 指标元数据中定义一个类型 ？
// 指标元数据类型 -> endpoint type
func buildEndpointType(item *model.MetricInfo) string {
	if item.EndpointType == "NETWORK" {
		return config.EndpointKeyNetwork
	} else if item.EndpointType == "HOST" || item.EndpointType == "INSTANCE" {
		if strings.HasPrefix(item.Metric, "container") || strings.HasPrefix(item.Metric, "docker") {
			return config.EndpointKeyDocker
		} else {
			return config.EndpointKeyPM
		}
	}
	return ""
}
