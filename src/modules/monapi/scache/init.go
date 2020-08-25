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
		item, exists := mcache.MonitorItemCache.Get(stra.Exprs[0].Metric)
		if !exists {
			logger.Errorf("stra %s metric %s is not in monitorItem cache", stra.Name, stra.Exprs[0].Metric)
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
			if hostE.Exists(e.Ident) || hostE.Exists(e.Alias) {
				stra.Endpoints = append(stra.Endpoints, e.Ident)
				continue
			}
			if !hostN.Exists(e.Ident) && !hostN.Exists(e.Alias) {
				stra.Endpoints = append(stra.Endpoints, e.Ident)
				continue
			}

			tags, err := dataobj.SplitTagsString(e.Tags)
			if err != nil {
				logger.Errorf("split endpoint %s tags %s error, %s", e.Ident, e.Tags, err)
				continue
			}
			// env filter
			if envTag, ok := tags["env"]; ok {
				if envE.Exists(envTag) {
					stra.Endpoints = append(stra.Endpoints, e.Ident)
					continue
				}
				if !envN.Exists(envTag) {
					stra.Endpoints = append(stra.Endpoints, e.Ident)
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
	//排除节点为空，直接将所有叶子节点返回
	if len(exclNid) == 0 {
		return ids, nil
	}

	exclLeafIds, err := GetExclLeafIds(exclNid)
	if err != nil {
		return leafIds, err
	}

	for _, id := range ids {
		idsMap[id] = true
	}
	for _, id := range exclLeafIds {
		delete(idsMap, id)
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
func buildEndpointType(item *model.MonitorItem) string {
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
