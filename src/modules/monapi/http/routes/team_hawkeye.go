package routes

import (
	"github.com/didi/nightingale/src/model"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"strconv"
	"strings"
)

func teamHawkeyeListGet(c *gin.Context) {
	limit := queryInt(c, "limit", 10000)
	query := queryStr(c, "query", "")
	nid := mustQueryInt64(c, "nid")
	edit := queryInt(c, "edit", 1)
	var nids []int64
	m := make(map[int64]string)

	if edit == 1 {
		srvTrees, err := model.SrvTreeDescendants(nid)
		if err != nil {

		}

		for _, srvTree := range srvTrees {
			nids = append(nids, srvTree.Id)
			m[srvTree.Id] = srvTree.NodeCode
		}
	} else {
		nids = append(nids, nid)
		srvTree, err := model.GetNodeById(nid)
		if err != nil {

		}
		m[srvTree.Id] = srvTree.Note
	}

	var ids []int64
	if query != "" {
		idsStrs := strings.Split(query, ",")
		for _, idStr := range idsStrs {
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
		team.NodeCode = m[nid]
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
	userIds, err := model.SaveSSOUser(f.UserNames)
	if err != nil {
		errors.Bomb("teamHawkeyeAddPost SaveSSOUser err,userNames:%v", f.UserNames)
	}
	renderMessage(c, model.TeamHawkeyeAdd(f.Ident, f.Name, 0, userIds, f.Nid))
}

func teamHawkeyePut(c *gin.Context) {
	me := loginUser(c)

	var f teamHawkeyeForm
	errors.Dangerous(c.ShouldBind(&f))

	userIds, err := model.SaveSSOUser(f.UserNames)
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
