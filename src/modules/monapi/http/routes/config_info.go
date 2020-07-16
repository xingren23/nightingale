package routes

import (
	"github.com/didi/nightingale/src/model"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"time"
)

func cfgListGet(c *gin.Context) {
	limit := queryInt(c, "limit", 20)
	query := queryStr(c, "query", "")

	total, err := model.ConfigInfoTotal(query)
	errors.Dangerous(err)
	list, err := model.ConfigInfoGets(query, limit, offset(c, limit, total))
	errors.Dangerous(err)

	renderData(c, gin.H{
		"list":  list,
		"total": total,
	}, nil)
}

type configInfoForm struct {
	CfgGroup string `json:"cfg_group" binding:"required"`
	CfgKey   string `json:"cfg_key" binding:"required"`
	CfgValue string `json:"cfg_value" binding:"required"`
}

func cfgPost(c *gin.Context) {
	me := loginUser(c)
	var cfg configInfoForm
	errors.Dangerous(c.ShouldBind(&cfg))

	configInfo := model.ConfigInfo{
		CfgGroup:   cfg.CfgGroup,
		CfgKey:     cfg.CfgKey,
		CfgValue:   cfg.CfgValue,
		CreateTime: time.Now(),
		CreateBy:   me.Id,
		UpdateTime: time.Now(),
		UpdateBy:   me.Id,
		Status:     1,
	}
	renderMessage(c, configInfo.Add())
}

func cfgPut(c *gin.Context) {
	me := loginUser(c)

	var f configInfoForm
	errors.Dangerous(c.ShouldBind(&f))

	cfg, err := model.ConfigInfoGet("id", urlParamInt64(c, "id"))
	errors.Dangerous(err)
	cfg.CfgGroup = f.CfgGroup
	cfg.CfgKey = f.CfgKey
	cfg.CfgValue = f.CfgValue
	cfg.UpdateTime = time.Now()
	cfg.UpdateBy = me.Id

	renderMessage(c, cfg.Update("cfg_group", "cfg_key", "cfg_value", "update_time", "update_by"))
}

func cfgDel(c *gin.Context) {
	id := urlParamInt64(c, "id")
	target, err := model.ConfigInfoGet("id", id)
	errors.Dangerous(err)

	if target == nil {
		renderMessage(c, nil)
		return
	}

	renderMessage(c, target.Del())
}

func cfgGet(c *gin.Context) {
	id := urlParamInt64(c, "id")
	target, err := model.ConfigInfoGet("id", id)
	errors.Dangerous(err)

	if target == nil {
		renderMessage(c, nil)
		return
	}
	renderData(c, target, err)
}

func collectGetSieve(c *gin.Context) {
	cfgs, err := model.ConfigInfoGetByQ("collector", "illegal_chars")
	renderData(c, cfgs, err)
}
