package routes

import (
	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/http/middleware"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/toolkits/pkg/errors"
	"strings"
)

func eventCurGetsHawkeye(c *gin.Context) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	user := middleware.GetUser(c)
	stime := queryInt64(c, "stime", 0)
	etime := queryInt64(c, "etime", 0)
	metric := queryStr(c, "metric", "")
	endpoint := queryStr(c, "endpoint", "")

	limit := queryInt(c, "limit", 20)

	priorities := queryStr(c, "priorities", "")
	sendtypes := queryStr(c, "sendtypes", "")

	total, err := model.EventCurTotalHawkeye(stime, etime, user.Id, metric, endpoint, strings.Split(priorities, ","), strings.Split(sendtypes, ","))
	errors.Dangerous(err)

	events, err := model.EventCurGetsHawkeye(stime, etime, user.Id, metric, endpoint, strings.Split(priorities, ","), strings.Split(sendtypes, ","), limit, offset(c, limit, total))
	errors.Dangerous(err)

	datList := []eventData{}
	for i := 0; i < len(events); i++ {
		users, err := model.UserNameGetByIds(events[i].Users)
		errors.Dangerous(err)

		groups, err := model.TeamNameGetsByIds(events[i].Groups)
		errors.Dangerous(err)

		claimants, err := model.UserNameGetByIds(events[i].Claimants)
		errors.Dangerous(err)

		var detail []model.EventDetail
		err = json.Unmarshal([]byte(events[i].Detail), &detail)
		errors.Dangerous(err)

		var tags string
		if len(detail) > 0 {
			tags = dataobj.SortedTags(detail[0].Tags)
		}

		alertUpgrade, err := model.EventAlertUpgradeUnMarshal(events[i].AlertUpgrade)
		errors.Dangerous(err)

		alertUsers, err := model.UserNameGetByIds(alertUpgrade.Users)
		errors.Dangerous(err)

		alertGroups, err := model.TeamNameGetsByIds(alertUpgrade.Groups)
		errors.Dangerous(err)

		dat := eventData{
			Id:          events[i].Id,
			Sid:         events[i].Sid,
			Sname:       events[i].Sname,
			NodePath:    events[i].NodePath,
			Endpoint:    events[i].Endpoint,
			Priority:    events[i].Priority,
			EventType:   events[i].EventType,
			Category:    events[i].Category,
			HashId:      events[i].HashId,
			Etime:       events[i].Etime,
			Value:       events[i].Value,
			Info:        events[i].Info,
			Tags:        tags,
			Created:     events[i].Created,
			Nid:         events[i].Nid,
			Users:       users,
			Groups:      groups,
			Detail:      detail,
			Status:      model.StatusConvert(model.GetStatusByFlag(events[i].Status)),
			Claimants:   claimants,
			NeedUpgrade: events[i].NeedUpgrade,
			AlertUpgrade: AlertUpgrade{
				Groups:   alertGroups,
				Users:    alertUsers,
				Duration: alertUpgrade.Duration,
				Level:    alertUpgrade.Level,
			},
		}

		datList = append(datList, dat)
	}

	renderData(c, map[string]interface{}{
		"total": total,
		"list":  datList,
	}, nil)
}

func eventHisGetsHawkeye(c *gin.Context) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	stime := mustQueryInt64(c, "stime")
	etime := mustQueryInt64(c, "etime")
	nid := mustQueryInt64(c, "nid")
	metric := queryStr(c, "metric", "")
	endpoints := queryStr(c, "endpoints", "")

	limit := queryInt(c, "limit", 20)

	priorities := queryStr(c, "priorities", "")
	sendtypes := queryStr(c, "sendtypes", "")
	eventType := queryStr(c, "type", "")

	total, err := model.EventTotalHawkeye(stime, etime, nid, metric, eventType, strings.Split(endpoints, ","), strings.Split(priorities, ","), strings.Split(sendtypes, ","))
	errors.Dangerous(err)

	events, err := model.EventGetsHawkeye(stime, etime, nid, metric, eventType, strings.Split(endpoints, ","), strings.Split(priorities, ","), strings.Split(sendtypes, ","), limit, offset(c, limit, total))
	errors.Dangerous(err)

	datList := []eventData{}
	for i := 0; i < len(events); i++ {
		users, err := model.UserNameGetByIds(events[i].Users)
		errors.Dangerous(err)

		groups, err := model.TeamNameGetsByIds(events[i].Groups)
		errors.Dangerous(err)

		var detail []model.EventDetail
		err = json.Unmarshal([]byte(events[i].Detail), &detail)
		errors.Dangerous(err)

		var tags string
		if len(detail) > 0 {
			tags = dataobj.SortedTags(detail[0].Tags)
		}

		alertUpgrade, err := model.EventAlertUpgradeUnMarshal(events[i].AlertUpgrade)
		errors.Dangerous(err)

		alertUsers, err := model.UserNameGetByIds(alertUpgrade.Users)
		errors.Dangerous(err)

		alertGroups, err := model.TeamNameGetsByIds(alertUpgrade.Groups)
		errors.Dangerous(err)

		dat := eventData{
			Id:          events[i].Id,
			Sid:         events[i].Sid,
			Sname:       events[i].Sname,
			NodePath:    events[i].NodePath,
			Endpoint:    events[i].Endpoint,
			Priority:    events[i].Priority,
			EventType:   events[i].EventType,
			Category:    events[i].Category,
			HashId:      events[i].HashId,
			Etime:       events[i].Etime,
			Value:       events[i].Value,
			Info:        events[i].Info,
			Tags:        tags,
			Created:     events[i].Created,
			Nid:         events[i].Nid,
			Users:       users,
			Groups:      groups,
			Detail:      detail,
			Status:      model.StatusConvert(model.GetStatusByFlag(events[i].Status)),
			NeedUpgrade: events[i].NeedUpgrade,
			AlertUpgrade: AlertUpgrade{
				Groups:   alertGroups,
				Users:    alertUsers,
				Duration: alertUpgrade.Duration,
				Level:    alertUpgrade.Level,
			},
		}

		datList = append(datList, dat)
	}

	renderData(c, map[string]interface{}{
		"total": total,
		"list":  datList,
	}, nil)
}
