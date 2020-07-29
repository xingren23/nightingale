package meicai

import (
	"fmt"
	"strings"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/str"
)

// InitNode 初始化服务树节点
func (m *Meicai) InitNode() {
	// get srvtree
	url := fmt.Sprintf("%s%s", m.OpsAddr, OpsSrvtreeRootPath)
	nodes, err := SrvTreeGets(url, m.Timeout)
	if err != nil {
		logger.Errorf("get srvtree failed, %s", err)
		return
	}
	// update local cache
	m.srvTreeCache.SetAll(nodes)
	logger.Info("srvtree node init done")
}

func (m *Meicai) NodeGets() (nodes []dataobj.Node, err error) {
	return nodes, err
}

func (m *Meicai) NodeGetsByPaths(paths []string) ([]dataobj.Node, error) {
	if len(paths) == 0 {
		return []dataobj.Node{}, nil
	}

	var nodes []dataobj.Node
	return nodes, nil
}

func (m *Meicai) NodeByIds(ids []int64) ([]dataobj.Node, error) {
	if len(ids) == 0 {
		return []dataobj.Node{}, nil
	}

	nodes := make([]dataobj.Node, 0)
	return nodes, nil
}

func (m *Meicai) NodeQueryPath(query string, limit int) (nodes []dataobj.Node, err error) {
	nodes = make([]dataobj.Node, 0)
	return nodes, nil
}

func (m *Meicai) TreeSearchByPath(query string) (nodes []dataobj.Node, err error) {
	nodes = make([]dataobj.Node, 0)
	return nodes, nil
}

func (m *Meicai) NodeGet(col string, val interface{}) (*dataobj.Node, error) {
	var obj dataobj.Node

	return &obj, nil
}

func (m *Meicai) NodesGetByIds(ids []int64) ([]dataobj.Node, error) {
	var objs []dataobj.Node
	return objs, nil
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

	ret := make([]int64, 0)
	return ret, nil
}

func (m *Meicai) Pids(n *dataobj.Node) ([]int64, error) {
	ret := make([]int64, 0)
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
