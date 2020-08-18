package routes

import (
	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/scache"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/logger"
	"strconv"
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
