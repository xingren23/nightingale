package routes

import (
	"github.com/didi/nightingale/src/model"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"strings"
)

func maskconfGetsHawkeye(c *gin.Context) {
	nid := urlParamInt64(c, "id")
	endpoints := queryStr(c, "endpoints", "")

	objs, err := model.MaskconfGetsHawkeye(nid, strings.Split(endpoints, ","))
	errors.Dangerous(err)

	for i := 0; i < len(objs); i++ {
		errors.Dangerous(objs[i].FillEndpoints())
	}

	renderData(c, objs, nil)
}
