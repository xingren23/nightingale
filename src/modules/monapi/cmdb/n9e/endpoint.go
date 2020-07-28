package n9e

import (
	"fmt"
	"strings"

	"xorm.io/xorm"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
	"github.com/toolkits/pkg/str"
)

func (c *N9e) EndpointGet(col string, val interface{}) (*dataobj.Endpoint, error) {
	var obj dataobj.Endpoint
	has, err := c.DB["mon"].Where(col+"=?", val).Get(&obj)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &obj, nil
}

func (c *N9e) Update(e *dataobj.Endpoint, cols ...string) error {
	_, err := c.DB["mon"].Where("id=?", e.Id).Cols(cols...).Update(e)
	return err
}

func (c *N9e) endpointTotal(query, batch, field string) (int64, error) {
	session := c.buildEndpointWhere(query, batch, field)
	return session.Count(new(dataobj.Endpoint))
}

func (c *N9e) EndpointGets(query, batch, field string, limit, offset int) ([]dataobj.Endpoint, int64, error) {
	var objs []dataobj.Endpoint
	total, err := c.endpointTotal(query, batch, field)
	if err != nil {
		return objs, total, err
	}
	if int64(offset) > total {
		return objs, total, fmt.Errorf("offset > total, %d > %d", offset, total)
	}
	session := c.buildEndpointWhere(query, batch, field).OrderBy(field).Limit(limit, offset)
	err = session.Find(&objs)
	return objs, total, err
}

func (c *N9e) buildEndpointWhere(query, batch, field string) *xorm.Session {
	session := c.DB["mon"].Table(new(dataobj.Endpoint))

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

func (c *N9e) EndpointImport(endpoints []string) error {
	count := len(endpoints)
	if count == 0 {
		return nil
	}

	session := c.DB["mon"].NewSession()
	defer session.Close()

	for i := 0; i < count; i++ {
		arr := strings.Split(endpoints[i], "::")

		ident := strings.TrimSpace(arr[0])
		alias := ""
		if len(arr) == 2 {
			alias = strings.TrimSpace(arr[1])
		}

		if ident == "" {
			continue
		}

		err := c.endpointImport(session, ident, alias)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *N9e) endpointImport(session *xorm.Session, ident, alias string) error {
	var endpoint dataobj.Endpoint
	has, err := session.Where("ident=?", ident).Get(&endpoint)
	if err != nil {
		return err
	}

	if has {
		if alias != "" {
			endpoint.Alias = alias
			_, err = session.Where("ident=?", ident).Cols("alias").Update(endpoint)
		}
	} else {
		_, err = session.Insert(dataobj.Endpoint{Ident: ident, Alias: alias})
	}

	return err
}

func (c *N9e) EndpointDel(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	bindings, err := c.nodeEndpointGetByEndpointIds(ids)
	if err != nil {
		return err
	}

	for i := 0; i < len(bindings); i++ {
		err = c.NodeEndpointUnbind(bindings[i].NodeId, bindings[i].EndpointId)
		if err != nil {
			return err
		}
	}

	if _, err := c.DB["mon"].In("id", ids).Delete(new(dataobj.Endpoint)); err != nil {
		return err
	}

	return nil
}

func (c *N9e) buildEndpointUnderNodeWhere(leafids []int64, query, batch, field string) *xorm.Session {
	session := c.DB["mon"].Where("id in (select endpoint_id from node_endpoint where node_id in (" + str.IdsString(leafids) + "))")

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

func (c *N9e) endpointUnderNodeTotal(leafids []int64, query, batch, field string) (int64, error) {
	session := c.buildEndpointUnderNodeWhere(leafids, query, batch, field)
	return session.Count(new(dataobj.Endpoint))
}

func (c *N9e) EndpointUnderNodeGets(leafids []int64, query, batch, field string, limit,
	offset int) ([]dataobj.Endpoint, int64, error) {
	var objs []dataobj.Endpoint
	total, err := c.endpointUnderNodeTotal(leafids, query, batch, field)
	if err != nil {
		return objs, total, err
	}
	if int64(offset) > total {
		return objs, total, fmt.Errorf("offset > total, %d > %d", offset, total)
	}

	session := c.buildEndpointUnderNodeWhere(leafids, query, batch, field).Limit(limit, offset).OrderBy(field)
	err = session.Find(&objs)
	return objs, total, err
}

func (c *N9e) EndpointIdsByIdents(idents []string) ([]int64, error) {
	idents = str.TrimStringSlice(idents)
	if len(idents) == 0 {
		return []int64{}, nil
	}

	var objs []dataobj.Endpoint
	err := c.DB["mon"].In("ident", idents).Find(&objs)
	if err != nil {
		return []int64{}, err
	}

	cnt := len(objs)
	ret := make([]int64, 0, cnt)
	for i := 0; i < cnt; i++ {
		ret = append(ret, objs[i].Id)
	}

	return ret, nil
}

func (c *N9e) EndpointBindings(endpointIds []int64) ([]dataobj.EndpointBinding, error) {
	var nes []NodeEndpoint
	err := c.DB["mon"].In("endpoint_id", endpointIds).Find(&nes)
	if err != nil {
		return []dataobj.EndpointBinding{}, err
	}

	cnt := len(nes)
	if cnt == 0 {
		return []dataobj.EndpointBinding{}, nil
	}

	h2n := make(map[int64][]int64)
	arr := make([]int64, 0, cnt)
	for i := 0; i < cnt; i++ {
		arr = append(arr, nes[i].EndpointId)
		h2n[nes[i].EndpointId] = append(h2n[nes[i].EndpointId], nes[i].NodeId)
	}

	var endpoints []dataobj.Endpoint
	err = c.DB["mon"].In("id", arr).Find(&endpoints)
	if err != nil {
		return []dataobj.EndpointBinding{}, err
	}

	cnt = len(endpoints)
	ret := make([]dataobj.EndpointBinding, 0, cnt)
	for i := 0; i < cnt; i++ {
		nodeids := h2n[endpoints[i].Id]
		if len(nodeids) == 0 {
			continue
		}

		var nodes []dataobj.Node
		err = c.DB["mon"].In("id", nodeids).Find(&nodes)
		if err != nil {
			return []dataobj.EndpointBinding{}, err
		}

		b := dataobj.EndpointBinding{
			Ident: endpoints[i].Ident,
			Alias: endpoints[i].Alias,
			Nodes: nodes,
		}

		ret = append(ret, b)
	}

	return ret, nil
}

func (c *N9e) EndpointUnderLeafs(leafIds []int64) ([]dataobj.Endpoint, error) {
	var endpoints []dataobj.Endpoint
	if len(leafIds) == 0 {
		return []dataobj.Endpoint{}, nil
	}

	err := c.DB["mon"].Where("id in (select endpoint_id from node_endpoint where node_id in (" + str.IdsString(leafIds) + "))").Find(&endpoints)
	return endpoints, err
}
