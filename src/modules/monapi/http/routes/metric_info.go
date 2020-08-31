package routes

import (
	"github.com/didi/nightingale/src/modules/monapi/mcache"
	"github.com/gin-gonic/gin"
)

func metricInfoGet(c *gin.Context) {
	renderData(c, mcache.MetricInfoCache.GetAll(), nil)
}
