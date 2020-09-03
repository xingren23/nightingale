package model

func MaskconfGetsHawkeye(nodeId int64) ([]Maskconf, error) {
	var relatedNodeIds []int64
	err := DB["mon"].Table("maskconf").Select("nid").Find(&relatedNodeIds)
	if err != nil {
		return nil, err
	}

	if len(relatedNodeIds) == 0 {
		return []Maskconf{}, nil
	}

	var objs []Maskconf
	err = DB["mon"].Where("nid = ?", nodeId).Find(&objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
}
