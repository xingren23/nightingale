package routes

import (
	"github.com/didi/nightingale/src/model"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
)

func maskconfGetsHawkeye(c *gin.Context) {
	nid := urlParamInt64(c, "id")

	objs, err := model.MaskconfGetsHawkeye(nid)
	errors.Dangerous(err)

	for i := 0; i < len(objs); i++ {
		errors.Dangerous(objs[i].FillEndpoints())
	}

	renderData(c, objs, nil)
}
