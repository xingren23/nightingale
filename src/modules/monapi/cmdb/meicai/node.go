package meicai

import (
	"fmt"
	"strings"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
	"github.com/toolkits/pkg/str"
)

func (m *Meicai) NodeGets() (nodes []dataobj.Node, err error) {
	nodes, err = m.nodeGetsWhere("")
	return nodes, err
}

func (m *Meicai) nodeGetsWhere(where string, args ...interface{}) (nodes []dataobj.Node, err error) {
	if where != "" {
		err = m.DB["mon"].Where(where, args...).Find(&nodes)
	} else {
		err = m.DB["mon"].Find(&nodes)
	}
	return nodes, err
}

func (m *Meicai) NodeGetsByPaths(paths []string) ([]dataobj.Node, error) {
	if len(paths) == 0 {
		return []dataobj.Node{}, nil
	}

	var nodes []dataobj.Node
	err := m.DB["mon"].In("path", paths).Find(&nodes)
	return nodes, err
}

func (m *Meicai) NodeByIds(ids []int64) ([]dataobj.Node, error) {
	if len(ids) == 0 {
		return []dataobj.Node{}, nil
	}
	var objs []dataobj.Node
	err := m.DB["mon"].In("id", ids).Find(&objs)
	return objs, err
}

func (m *Meicai) NodeQueryPath(query string, limit int) (nodes []dataobj.Node, err error) {
	err = m.DB["mon"].Where("path like ?", "%"+query+"%").OrderBy("path").Limit(limit).Find(&nodes)
	return nodes, err
}

func (m *Meicai) TreeSearchByPath(query string) (nodes []dataobj.Node, err error) {
	session := m.DB["mon"].NewSession()
	defer session.Close()

	if strings.Contains(query, " ") {
		arr := strings.Fields(query)
		cnt := len(arr)
		for i := 0; i < cnt; i++ {
			session.Where("path like ?", "%"+arr[i]+"%")
		}
		err = session.Find(&nodes)
	} else {
		err = session.Where("path like ?", "%"+query+"%").Find(&nodes)
	}

	if err != nil {
		return
	}

	cnt := len(nodes)
	if cnt == 0 {
		return
	}

	pathset := make(map[string]struct{})
	for i := 0; i < cnt; i++ {
		pathset[nodes[i].Path] = struct{}{}

		paths := dataobj.Paths(nodes[i].Path)
		for j := 0; j < len(paths); j++ {
			pathset[paths[j]] = struct{}{}
		}
	}

	var objs []dataobj.Node
	err = session.In("path", str.MtoL(pathset)).Find(&objs)
	return objs, err
}

func (m *Meicai) NodeGet(col string, val interface{}) (*dataobj.Node, error) {
	var obj dataobj.Node
	has, err := m.DB["mon"].Where(col+"=?", val).Get(&obj)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &obj, nil
}

func (m *Meicai) NodeValid(name, path string) error {
	if len(name) > 32 {
		return fmt.Errorf("name too long")
	}

	if len(path) > 255 {
		return fmt.Errorf("path too long")
	}

	if !str.IsMatch(name, `^[a-z0-9\-]+$`) {
		return fmt.Errorf("name permissible characters: [a-z0-9] and -")
	}

	arr := strings.Split(path, ".")
	if name != arr[len(arr)-1] {
		return fmt.Errorf("name and path not match")
	}

	return nil
}

func (m *Meicai) LeafIds(n *dataobj.Node) ([]int64, error) {
	if n.Leaf == 1 {
		return []int64{n.Id}, nil
	}

	var nodes []dataobj.Node
	// todo : 与夜莺逻辑不同
	err := m.DB["mon"].Where("path like ? and leaf=1", "%"+n.Path+"%").Find(&nodes)
	if err != nil {
		return []int64{}, err
	}

	cnt := len(nodes)
	arr := make([]int64, 0, cnt)
	for i := 0; i < cnt; i++ {
		arr = append(arr, nodes[i].Id)
	}

	return arr, nil
}

func (m *Meicai) Pids(n *dataobj.Node) ([]int64, error) {
	if n.Pid == 0 {
		return []int64{n.Pid}, nil
	}

	var objs []dataobj.Node
	arr := []int64{}
	paths := []string{}

	// corp.spruce,_owt.operations-center,_pdl.arch,_sg.hawkeye,_srv.web
	nodes := strings.Split(n.Path, "_")
	cnt := len(nodes)

	for i := 1; i < cnt; i++ {
		path := strings.Join(nodes[:cnt-i], "_")
		paths = append(paths, path)
	}

	err := m.DB["mon"].In("path", paths).Find(&objs)
	if err != nil {
		return []int64{}, err
	}

	cnt = len(objs)
	for i := 0; i < cnt; i++ {
		arr = append(arr, objs[i].Id)
	}

	return arr, nil
}

func (m *Meicai) CreateChild(n *dataobj.Node, name string, leaf int, note string) (int64, error) {
	return -1, fmt.Errorf("meicai cmdb not impliement %s interface", "CreateChild")
}

func (m *Meicai) Bind(n *dataobj.Node, endpointIds []int64, delOld int) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "Bind")
}

func (m *Meicai) Unbind(n *dataobj.Node, hostIds []int64) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "Unbind")
}

func (m *Meicai) Rename(n *dataobj.Node, name string) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "Rename")
}

func (m *Meicai) Del(n *dataobj.Node) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "Del")
}
