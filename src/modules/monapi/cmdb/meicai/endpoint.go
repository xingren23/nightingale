package meicai

import (
	"fmt"
	"strings"

	"github.com/toolkits/pkg/str"
	"xorm.io/xorm"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
)

// support ident col
func (m *Meicai) EndpointGet(col string, val interface{}) (*dataobj.Endpoint, error) {
	var obj dataobj.Endpoint
	has, err := m.DB["mon"].Where(col+"=?", val).Get(&obj)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &obj, nil
}

func (m *Meicai) EndpointGets(query, batch, field string, limit, offset int) ([]dataobj.Endpoint, int64, error) {
	var objs []dataobj.Endpoint
	total, err := m.endpointTotal(query, batch, field)
	if err != nil {
		return objs, total, err
	}
	if int64(offset) > total {
		return objs, total, fmt.Errorf("offset > total, %d > %d", offset, total)
	}
	session := m.buildEndpointWhere(query, batch, field).OrderBy(field).Limit(limit, offset)
	err = session.Find(&objs)
	return objs, total, err
}

func (m *Meicai) endpointTotal(query, batch, field string) (int64, error) {
	session := m.buildEndpointWhere(query, batch, field)
	return session.Count(new(dataobj.Endpoint))
}

func (m *Meicai) buildEndpointWhere(query, batch, field string) *xorm.Session {
	session := m.DB["mon"].Table(new(dataobj.Endpoint))

	if batch == "" && query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			q := "%" + arr[i] + "%"
			session = session.Where("ident like ? or alias like ?", q, q)
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

func (m *Meicai) buildEndpointUnderNodeWhere(leafids []int64, query, batch, field string) *xorm.Session {
	session := m.DB["mon"].Where("id in (select endpoint_id from node_endpoint where node_id in (" + str.IdsString(leafids) + "))")

	if batch == "" && query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			q := "%" + arr[i] + "%"
			session = session.Where("ident like ? or alias like ?", q, q)
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

func (m *Meicai) endpointUnderNodeTotal(leafids []int64, query, batch, field string) (int64, error) {
	session := m.buildEndpointUnderNodeWhere(leafids, query, batch, field)
	return session.Count(new(dataobj.Endpoint))
}

func (m *Meicai) EndpointUnderNodeGets(leafids []int64, query, batch, field string, limit, offset int) ([]dataobj.Endpoint, int64, error) {
	var objs []dataobj.Endpoint
	total, err := m.endpointUnderNodeTotal(leafids, query, batch, field)
	if err != nil {
		return objs, total, err
	}
	if int64(offset) > total {
		return objs, total, fmt.Errorf("offset > total, %d > %d", offset, total)
	}

	session := m.buildEndpointUnderNodeWhere(leafids, query, batch, field).Limit(limit, offset).OrderBy(field)
	err = session.Find(&objs)
	return objs, total, err
}

func (m *Meicai) EndpointIdsByIdents(idents []string) ([]int64, error) {
	return []int64{}, fmt.Errorf("meicai cmdb not impliement %s interface", "EndpointIdsByIdents")
}

func (m *Meicai) EndpointBindings(endpointIds []int64) ([]dataobj.EndpointBinding, error) {

	ret := make([]dataobj.EndpointBinding, 0)
	return ret, nil
}

func (m *Meicai) EndpointUnderLeafs(leafIds []int64) ([]dataobj.Endpoint, error) {
	var endpoints []dataobj.Endpoint
	if len(leafIds) == 0 {
		return []dataobj.Endpoint{}, nil
	}

	err := m.DB["mon"].Where("id in (select endpoint_id from node_endpoint where node_id in (" + str.IdsString(
		leafIds) + "))").Find(&endpoints)
	return endpoints, err
}

func (m *Meicai) Update(e *dataobj.Endpoint, cols ...string) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "Updata")
}

func (m *Meicai) EndpointImport(endpoints []string) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "EndpointImport")
}

func (m *Meicai) EndpointDel(ids []int64) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "EndpointDel")
}
