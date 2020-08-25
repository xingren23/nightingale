package meicai

import (
	"fmt"
	"strings"

	"github.com/toolkits/pkg/str"
	"xorm.io/xorm"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
)

func (m *Meicai) AppInstanceGets(query, batch, field string, limit, offset int) ([]dataobj.AppInstance, int64, error) {
	var objs []dataobj.AppInstance
	total, err := m.appInstanceTotal(query, batch, field)
	if err != nil {
		return objs, total, err
	}
	if int64(offset) > total {
		return objs, total, fmt.Errorf("offset > total, %d > %d", offset, total)
	}
	session := m.buildAppInstanceWhere(query, batch, field).OrderBy(field).Limit(limit, offset)
	err = session.Find(&objs)
	return objs, total, err
}

func (m *Meicai) AppInstanceUnderLeafs(leafIds []int64) ([]dataobj.AppInstance, error) {
	var objs []dataobj.AppInstance
	if len(leafIds) == 0 {
		return []dataobj.AppInstance{}, nil
	}

	err := m.DB["mon"].Where("node_id in (" + str.IdsString(leafIds) + ")").Find(&objs)
	return objs, err
}

func (m *Meicai) appInstanceTotal(query, batch, field string) (int64, error) {
	session := m.buildAppInstanceWhere(query, batch, field)
	return session.Count(new(dataobj.AppInstance))
}

func (m *Meicai) buildAppInstanceWhere(query, batch, field string) *xorm.Session {
	session := m.DB["mon"].Table(new(dataobj.AppInstance))

	if batch == "" && query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			q := "%" + arr[i] + "%"
			session = session.Where("ident like ? or app like ?", q, q)
		}
	}

	if batch != "" {
		endpoints := str.ParseCommaTrim(batch)
		if len(endpoints) > 0 {
			session = session.In(field, endpoints)
		}
	}
	return session
}
