package routes

import (
	"strconv"
	"strings"

	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/auth/meicai"
	"github.com/didi/nightingale/src/modules/monapi/cmdb"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
)

func teamHawkeyeListGet(c *gin.Context) {
	limit := queryInt(c, "limit", 10000)
	query := queryStr(c, "query", "")
	nid := mustQueryInt64(c, "nid")
	edit := queryInt(c, "edit", 1)
	var nids []int64
	m := make(map[int64]string)

	curNode, err := cmdb.GetCmdb().NodeGet("id", nid)
	errors.Dangerous(err)
	nids = append(nids, nid)
	m[curNode.Id] = curNode.Name

	if edit == 1 {
		pids, err := cmdb.GetCmdb().Pids(curNode)
		errors.Dangerous(err)

		pNodes, err := cmdb.GetCmdb().NodeByIds(pids)
		for _, node := range pNodes {
			nids = append(nids, node.Id)
			m[node.Id] = node.Name
		}
	}

	// 策略编辑id反查team
	var ids []int64
	if query != "" {
		for _, idStr := range strings.Split(query, ",") {
			id, _ := strconv.ParseInt(idStr, 10, 64)
			ids = append(ids, id)
		}
	}
	total, err := model.TeamHawkeyeTotal(nids, ids)
	errors.Dangerous(err)

	list, err := model.TeamHawkeyeGets(nids, ids, limit, offset(c, limit, total))
	errors.Dangerous(err)

	for i := 0; i < len(list); i++ {
		errors.Dangerous(list[i].FillObjs())
	}

	for _, team := range list {
		team.NodeCode = m[team.Nid]
	}

	renderData(c, gin.H{
		"list":  list,
		"total": total,
	}, nil)
}

type teamHawkeyeForm struct {
	Ident     string   `json:"ident"`
	Name      string   `json:"name"`
	Mgmt      int      `json:"mgmt"`
	Admins    []int64  `json:"admins"`
	Nid       int64    `json:"nid"`
	UserNames []string `json:"userNames"`
}

func teamHawkeyeAddPost(c *gin.Context) {
	var f teamHawkeyeForm
	errors.Dangerous(c.ShouldBind(&f))
	userIds, err := meicai.SaveSSOUser(f.UserNames)
	if err != nil {
		errors.Bomb("save user fail, err:[%s], ", err)
	}
	renderMessage(c, model.TeamHawkeyeAdd(f.Ident, f.Name, 0, userIds, f.Nid))
}

func teamHawkeyePut(c *gin.Context) {
	me := loginUser(c)

	var f teamHawkeyeForm
	errors.Dangerous(c.ShouldBind(&f))

	userIds, err := meicai.SaveSSOUser(f.UserNames)
	if err != nil {
		errors.Bomb("teamHawkeyePut SaveSSOUser err,userNames:%v", f.UserNames)
	}

	t, err := model.TeamGet("id", urlParamInt64(c, "id"))
	errors.Dangerous(err)

	if t == nil {
		errors.Bomb("no such team")
	}

	can, err := me.CanModifyTeam(t)
	errors.Dangerous(err)
	if !can {
		errors.Bomb("no privilege")
	}

	renderMessage(c, t.ModifyHawkeye(f.Ident, f.Name, f.Mgmt, f.Admins, userIds))
}
