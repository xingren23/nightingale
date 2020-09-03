package model

func MaskconfGetsHawkeye(nodeId int64, endpoints []string) ([]Maskconf, error) {
	var relatedNodeIds []int64
	err := DB["mon"].Table("maskconf").Select("nid").Find(&relatedNodeIds)
	if err != nil {
		return nil, err
	}

	if len(relatedNodeIds) == 0 {
		return []Maskconf{}, nil
	}

	var objs []Maskconf
	session := DB["mon"].Where("nid = ?", nodeId)

	if len(endpoints) > 0 && endpoints[0] != "" {
		var maskIds []int64
		DB["mon"].Table("maskconf_endpoints").Select("mask_id").Find(&maskIds)
		session = session.In("id", maskIds)
	}

	err = session.Find(&objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
}
