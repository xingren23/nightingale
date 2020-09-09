package routes

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

// Config routes
func Config(r *gin.Engine) {
	sys := r.Group("/api/collector")
	{
		sys.GET("/ping", ping)
		sys.GET("/pid", pid)
		sys.GET("/addr", addr)

		sys.GET("/metric_info", getMetricInfo)
		sys.GET("/app_instances", getAppInstance)
		sys.GET("/endpoints", getEndpoint)
		sys.GET("/garbage", getGarbage)

		sys.GET("/stra", getStrategy)
		sys.GET("/cached", getLogCached)
		sys.POST("/push", pushData)
	}
	r.POST("/v1/push", pushDataV1)

	pprof.Register(r, "/api/collector/debug/pprof")
}
