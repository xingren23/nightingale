package meicai

import (
	"fmt"
	"strings"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/str"
)

// InitNode 初始化服务树节点
func (m *Meicai) InitNode() error {
	// get srvtree
	url := fmt.Sprintf("%s%s", m.OpsAddr, OpsSrvtreeRootPath)
	nodes, err := SrvTreeGets(url, m.Timeout)
	if err != nil {
		logger.Errorf("get srvtree failed, %s", err)
		return err
	}
	// update local cache
	m.srvTreeCache.SetAll(nodes)
	logger.Info("srvtree node init done")
	return nil
}

func (m *Meicai) NodeGets() (nodes []dataobj.Node, err error) {
	nodes = m.srvTreeCache.GetNodes()
	if len(nodes) == 0 {
		err = fmt.Errorf("nodes is empty")
	}
	return nodes, err
}

func (m *Meicai) NodeGetsByPaths(paths []string) (nodes []dataobj.Node, err error) {
	if len(paths) == 0 {
		return []dataobj.Node{}, nil
	}

	pathNodes := m.srvTreeCache.GetPathNodes()
	nodes = make([]dataobj.Node, 0)
	for _, path := range paths {
		if node, exist := pathNodes[path]; exist {
			nodes = append(nodes, *node)
		}
	}
	if len(nodes) == 0 {
		err = fmt.Errorf("could not find nodes, paths %v", paths)
	}

	return nodes, err
}

func (m *Meicai) NodeByIds(ids []int64) (nodes []dataobj.Node, err error) {
	if len(ids) == 0 {
		return []dataobj.Node{}, nil
	}

	idNodes := m.srvTreeCache.GetIdNodes()
	nodes = make([]dataobj.Node, 0)
	for _, id := range ids {
		if node, exist := idNodes[id]; exist {
			nodes = append(nodes, *node)
		}
	}
	if len(nodes) == 0 {
		err = fmt.Errorf("could not find nodes, ids %v", ids)
	}
	return nodes, err
}

func (m *Meicai) NodeQueryPath(query string, limit int) (nodes []dataobj.Node, err error) {
	pathNodes := m.srvTreeCache.GetPathNodes()
	nodes = make([]dataobj.Node, 0)
	for path, node := range pathNodes {
		if strings.Contains(path, query) {
			nodes = append(nodes, *node)
		}
		if len(nodes) == limit {
			logger.Infof("exceed query path %s limit %d", query, limit)
			break
		}
	}
	if len(nodes) == 0 {
		err = fmt.Errorf("could not find nodes, query %v", query)
	}

	return nodes, err
}

func (m *Meicai) TreeSearchByPath(query string) (nodes []dataobj.Node, err error) {
	pathNodes := m.srvTreeCache.GetPathNodes()
	nodes = make([]dataobj.Node, 0)
	for path, node := range pathNodes {
		if strings.Contains(path, query) {
			nodes = append(nodes, *node)
		}
	}
	if len(nodes) == 0 {
		err = fmt.Errorf("could not find nodes, query %v", query)
	}
	return nodes, err
}

func (m *Meicai) NodeGet(col string, val interface{}) (*dataobj.Node, error) {
	var obj dataobj.Node
	nodes := m.srvTreeCache.GetNodes()
	for _, node := range nodes {
		if col == "id" && node.Id == val {
			obj = node
		} else if col == "name" && node.Name == val {
			obj = node
		} else if col == "path" && node.Path == val {
			obj = node
		}
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

// 包括自身节点
func (m *Meicai) LeafIds(n *dataobj.Node) ([]int64, error) {
	ret := make([]int64, 0)
	pathNodes := m.srvTreeCache.GetPathNodes()
	for path, node := range pathNodes {
		if strings.Contains(path, n.Path) && strings.Contains(n.Path, "_srv.") {
			ret = append(ret, node.Id)
		}
	}
	if len(ret) == 0 {
		return ret, fmt.Errorf("no leaf node, nid %d path %s", n.Id, n.Path)
	}
	return ret, nil
}

func (m *Meicai) Pids(n *dataobj.Node) ([]int64, error) {
	ret := make([]int64, 0)
	pathNodes := m.srvTreeCache.GetPathNodes()
	for path, node := range pathNodes {
		if strings.Contains(n.Path, path) && path != n.Path {
			ret = append(ret, node.Id)
		}
	}
	if len(ret) == 0 {
		return ret, fmt.Errorf("root node, nid %d path %s", n.Id, n.Path)
	}
	return ret, nil
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
