package n9e

import (
	"fmt"
)

type NodeEndpoint struct {
	NodeId     int64 `xorm:"'node_id'"`
	EndpointId int64 `xorm:"'endpoint_id'"`
}

func (NodeEndpoint) TableName() string {
	return "node_endpoint"
}

func (c *N9e) NodeIdsGetByEndpointId(endpointId int64) ([]int64, error) {
	if endpointId == 0 {
		return []int64{}, nil
	}

	var ids []int64
	err := c.DB["mon"].Table("node_endpoint").Where("endpoint_id = ?", endpointId).Select("node_id").Find(&ids)
	return ids, err
}

func (c *N9e) EndpointIdsByNodeIds(nodeIds []int64) ([]int64, error) {
	if len(nodeIds) == 0 {
		return []int64{}, nil
	}

	var ids []int64
	err := c.DB["mon"].Table("node_endpoint").In("node_id", nodeIds).Select("endpoint_id").Find(&ids)
	return ids, err
}

func (c *N9e) nodeEndpointGetByEndpointIds(endpointsIds []int64) ([]NodeEndpoint, error) {
	if len(endpointsIds) == 0 {
		return []NodeEndpoint{}, nil
	}

	var objs []NodeEndpoint
	err := c.DB["mon"].In("endpoint_id", endpointsIds).Find(&objs)
	return objs, err
}

func (c *N9e) nodeEndpointGetByNodeIds(nodeIds []int64) ([]NodeEndpoint, error) {
	if len(nodeIds) == 0 {
		return []NodeEndpoint{}, nil
	}

	var objs []NodeEndpoint
	err := c.DB["mon"].In("node_id", nodeIds).Find(&objs)
	return objs, err
}

func (c *N9e) NodeEndpointUnbind(nid, eid int64) error {
	_, err := c.DB["mon"].Where("node_id=? and endpoint_id=?", nid, eid).Delete(new(NodeEndpoint))
	return err
}

func (c *N9e) NodeEndpointBind(nid, eid int64) error {
	total, err := c.DB["mon"].Where("node_id=? and endpoint_id=?", nid, eid).Count(new(NodeEndpoint))
	if err != nil {
		return err
	}

	if total > 0 {
		return nil
	}

	endpoint, err := c.EndpointGet("id", eid)
	if err != nil {
		return err
	}

	if endpoint == nil {
		return fmt.Errorf("endpoint[id:%d] not found", eid)
	}

	_, err = c.DB["mon"].Insert(&NodeEndpoint{
		NodeId:     nid,
		EndpointId: eid,
	})

	return err
}
