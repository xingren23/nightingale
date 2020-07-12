package model

func EventCurTotalHawkeye(stime, etime int64, userId int64, metric, endpoint string, priorities, sendTypes []string) (int64, error) {
	session := DB["mon"].Where("FIND_IN_SET( ? ,alert_users) and ignore_alert=0", userId)

	if stime != 0 {
		session = DB["mon"].Where("etime > ?", stime)
	}

	if etime != 0 {
		session = DB["mon"].Where("etime < ?", etime)
	}

	if metric != "" {
		session = session.Where("value like ?", "%"+metric+"%")
	}

	if endpoint != "" {
		session = session.Where("endpoint like ?", "%"+endpoint+"%")
	}

	if len(priorities) > 0 && priorities[0] != "" {
		session = session.In("priority", priorities)
	}

	if len(sendTypes) > 0 && sendTypes[0] != "" {
		session = session.In("status", GetFlagsByStatus(sendTypes))
	}

	total, err := session.Count(new(EventCur))
	return total, err
}

func EventCurGetsHawkeye(stime, etime int64, userId int64, metric, endpoint string, priorities, sendTypes []string, limit, offset int) ([]EventCur, error) {
	var obj []EventCur

	session := DB["mon"].Where("FIND_IN_SET( ? ,alert_users) and ignore_alert=0", userId)

	if stime != 0 {
		session = DB["mon"].Where("etime > ?", stime)
	}

	if etime != 0 {
		session = DB["mon"].Where("etime < ?", etime)
	}

	if metric != "" {
		session = session.Where("value like ?", "%"+metric+"%")
	}

	if endpoint != "" {
		session = session.Where("endpoint like ?", "%"+endpoint+"%")
	}

	if len(priorities) > 0 && priorities[0] != "" {
		session = session.In("priority", priorities)
	}

	if len(sendTypes) > 0 && sendTypes[0] != "" {
		session = session.In("status", GetFlagsByStatus(sendTypes))
	}

	err := session.Desc("etime").Limit(limit, offset).Find(&obj)
	return obj, err
}
