package model

func EventTotalHawkeye(stime, etime, nid int64, metric, eventType string, endpoints, priorities, sendTypes []string) (int64, error) {
	session := DB["mon"].Where("etime > ? and etime < ? and nid = ?", stime, etime, nid)
	if len(priorities) > 0 && priorities[0] != "" {
		session = session.In("priority", priorities)
	}

	if len(sendTypes) > 0 && sendTypes[0] != "" {
		session = session.In("status", GetFlagsByStatus(sendTypes))
	}

	if eventType != "" {
		session = session.Where("event_type=?", eventType)
	}

	if metric != "" {
		session = session.Where("value like ?", "%"+metric+"%")
	}

	if len(endpoints) > 0 && endpoints[0] != "" {
		session = session.In("endpoint", endpoints)
	}

	total, err := session.Count(new(Event))
	return total, err
}

func EventGetsHawkeye(stime, etime, nid int64, metric, eventType string, endpoints, priorities, sendTypes []string, limit, offset int) ([]Event, error) {
	var objs []Event

	session := DB["mon"].Where("etime > ? and etime < ? and nid = ?", stime, etime, nid)
	if len(priorities) > 0 && priorities[0] != "" {
		session = session.In("priority", priorities)
	}

	if len(sendTypes) > 0 && sendTypes[0] != "" {
		session = session.In("status", GetFlagsByStatus(sendTypes))
	}

	if eventType != "" {
		session = session.Where("event_type=?", eventType)
	}

	if metric != "" {
		session = session.Where("value like ?", "%"+metric+"%")
	}

	if len(endpoints) > 0 && endpoints[0] != "" {
		session = session.In("endpoint", endpoints)
	}

	err := session.Desc("etime").Limit(limit, offset).Find(&objs)

	return objs, err
}
