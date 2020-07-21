package routes

import (
	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/didi/nightingale/src/modules/monapi/ecache"
	"github.com/didi/nightingale/src/modules/monapi/mcache"
	"github.com/didi/nightingale/src/modules/monapi/scache"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
)

type resourceForm struct {
	Nid    int64  `json:"nid"`
	Metric string `json:"metric"`
}

func resourcePost(c *gin.Context) {
	var f resourceForm
	errors.Dangerous(c.ShouldBind(&f))

	nodePath, exists := ecache.SrvTreeCache.Get(f.Nid)
	if !exists {
		errors.Bomb("nodePath is not exists: srvTreeId:%v", f.Nid)
	}

	srvTypes := make([]string, 0)
	if f.Metric == "" {
		srvTypes = append(srvTypes, config.EndpointKeyNetwork)
		srvTypes = append(srvTypes, config.EndpointKeyPM)
		srvTypes = append(srvTypes, config.EndpointKeyDocker)
	} else {
		//获取MonitorItem的类型
		item, exists := mcache.MonitorItemCache.Get(f.Metric)
		if !exists {
			errors.Bomb("MonitorItem is not exists: metric:%v", f.Metric)
		}

		srvType := scache.BuildSrvType(item)
		if srvType == "" {
			errors.Bomb("MonitorItem buildSrvType error: metric:%v", item.Metric)
		}
		srvTypes = append(srvTypes, srvType)
	}

	list := make([]data, 0)
	for _, srvType := range srvTypes {
		tagEndpoints, err := ecache.GetEndpointsFromRedis(srvType, nodePath)
		if err != nil {
			errors.Bomb("endpoints is not exists: nodePath:%v, srvType:%v, err:%v", nodePath, srvType, err)
		}
		for _, tagEndpoint := range tagEndpoints {
			list = append(list, data{Ip: tagEndpoint.Ip, Type: getType(srvType)})
		}
	}

	renderData(c, list, nil)
}

func appGet(c *gin.Context) {
	renderData(c, ecache.AppCache.GetAll(), nil)
}

func hostGet(c *gin.Context) {
	renderData(c, ecache.HostCache.GetAll(), nil)
}

func instanceGet(c *gin.Context) {
	renderData(c, ecache.InstanceCache.GetAll(), nil)
}

func networkGet(c *gin.Context) {
	renderData(c, ecache.NetworkCache.GetAll(), nil)
}

func monitorItemGet(c *gin.Context) {
	renderData(c, mcache.MonitorItemCache.GetAll(), nil)
}

func getType(srvType string) string {
	if srvType == config.EndpointKeyDocker {
		return "容器"
	}
	if srvType == config.EndpointKeyPM {
		return "物理机"
	}
	if srvType == config.EndpointKeyNetwork {
		return "网络设备"
	}

	return ""
}

type data struct {
	Ip   string `json:"ip"`
	Type string `json:"type"`
}
