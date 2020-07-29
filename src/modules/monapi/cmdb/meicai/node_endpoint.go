package meicai

import "fmt"

func (m *Meicai) NodeIdsGetByEndpointId(endpointId int64) ([]int64, error) {
	if endpointId == 0 {
		return []int64{}, nil
	}

	var ids []int64
	return ids, nil
}

func (m *Meicai) EndpointIdsByNodeIds(nodeIds []int64) ([]int64, error) {
	if len(nodeIds) == 0 {
		return []int64{}, nil
	}

	var ids []int64
	return ids, nil
}

func (m *Meicai) NodeEndpointUnbind(nid, eid int64) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "NodeEndpointUnbind")
}

func (m *Meicai) NodeEndpointBind(nid, eid int64) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "NodeEndpointBind")
}
