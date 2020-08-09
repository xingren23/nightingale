package meicai

import "fmt"

type NodeEndpoint struct {
	NodeId     int64 `xorm:"'node_id'"`
	EndpointId int64 `xorm:"'endpoint_id'"`
}

func (NodeEndpoint) TableName() string {
	return "node_endpoint"
}

func (m *Meicai) NodeIdsGetByEndpointId(endpointId int64) ([]int64, error) {
	if endpointId == 0 {
		return []int64{}, nil
	}

	var ids []int64
	err := m.DB["mon"].Table("node_endpoint").Where("endpoint_id = ?", endpointId).Select("node_id").Find(&ids)
	return ids, err
}

func (m *Meicai) EndpointIdsByNodeIds(nodeIds []int64) ([]int64, error) {
	if len(nodeIds) == 0 {
		return []int64{}, nil
	}

	var ids []int64
	err := m.DB["mon"].Table("node_endpoint").In("node_id", nodeIds).Select("endpoint_id").Find(&ids)
	return ids, err
}

func (m *Meicai) nodeEndpointGetByEndpointIds(endpointsIds []int64) ([]NodeEndpoint, error) {
	if len(endpointsIds) == 0 {
		return []NodeEndpoint{}, nil
	}

	var objs []NodeEndpoint
	err := m.DB["mon"].In("endpoint_id", endpointsIds).Find(&objs)
	return objs, err
}

func (m *Meicai) nodeEndpointGetByNodeIds(nodeIds []int64) ([]NodeEndpoint, error) {
	if len(nodeIds) == 0 {
		return []NodeEndpoint{}, nil
	}

	var objs []NodeEndpoint
	err := m.DB["mon"].In("node_id", nodeIds).Find(&objs)
	return objs, err
}

func (m *Meicai) NodeEndpointUnbind(nid, eid int64) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "NodeEndpointUnbind")
}

func (m *Meicai) NodeEndpointBind(nid, eid int64) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "NodeEndpointBind")
}
