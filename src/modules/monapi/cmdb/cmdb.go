package cmdb

import (
	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
	"github.com/toolkits/pkg/logger"
)

type IEndpointReadable interface {
	// 查询endpoint（col: id、ident or alias）
	EndpointGet(col string, val interface{}) (*dataobj.Endpoint, error)
	// 批量查询endpoint（query:ident or alias; field in batch; offset），支持翻页
	EndpointGets(query, batch, field string, limit, offset int) ([]dataobj.Endpoint, int64, error)
	// 批量查询endpoint（ident）
	EndpointIdsByIdents(idents []string) ([]int64, error)
	// 查询指定node下的endpoint（query:ident or alias; field in batch; offset）
	EndpointUnderNodeGets(leafids []int64, query, batch, field string, limit, offset int) ([]dataobj.Endpoint,
		int64, error)
	// 查询指定node下所有endpoint
	EndpointUnderLeafs(leafIds []int64) ([]dataobj.Endpoint, error)
	// 查询endpoint的bind信息
	EndpointBindings(endpointIds []int64) ([]dataobj.EndpointBinding, error)
}

type IEndpoint interface {
	IEndpointReadable

	Update(e *dataobj.Endpoint, cols ...string) error
	EndpointDel(ids []int64) error
	EndpointImport(endpoints []string) error
}

type INodeReadable interface {
	// 查询所有节点
	NodeGets() (nodes []dataobj.Node, err error)
	// 根据path查询
	NodeGetsByPaths(paths []string) ([]dataobj.Node, error)
	// 根据ids查询
	NodeByIds(ids []int64) ([]dataobj.Node, error)

	// 查询子节点id
	LeafIds(n *dataobj.Node) ([]int64, error)
	// 查询父节点id
	Pids(n *dataobj.Node) ([]int64, error)

	// 查询节点信息
	NodeGet(col string, val interface{}) (*dataobj.Node, error)
	// 根据path搜索节点
	NodeQueryPath(query string, limit int) (nodes []dataobj.Node, err error)
	// 服务树搜索（支持多查询条件）
	TreeSearchByPath(query string) (nodes []dataobj.Node, err error)
}

type INode interface {
	INodeReadable

	InitNode() error
	CreateChild(n *dataobj.Node, name string, leaf int, note string) (int64, error)
	Bind(n *dataobj.Node, endpointIds []int64, delOld int) error
	Unbind(n *dataobj.Node, hostIds []int64) error
	Del(n *dataobj.Node) error
	Rename(n *dataobj.Node, name string) error
}

type INodeEndpointReadable interface {
	// 查询endpoint所属的node
	NodeIdsGetByEndpointId(endpointId int64) ([]int64, error)
	// 查询node下的endpoint
	EndpointIdsByNodeIds(nodeIds []int64) ([]int64, error)
}

type INodeEndpoint interface {
	INodeEndpointReadable

	NodeEndpointUnbind(nid, eid int64) error
	NodeEndpointBind(nid, eid int64) error
}

type ICmdbReadable interface {
	IEndpointReadable
	INodeReadable
	INodeEndpointReadable
}
type ICmdb interface {
	IEndpoint
	INode
	INodeEndpoint
}

var defaultCmdb string
var registryCmdb map[string]ICmdb

func init() {
	registryCmdb = make(map[string]ICmdb, 0)
}

func GetCmdb() ICmdb {
	if cmdb, exists := registryCmdb[defaultCmdb]; exists {
		return cmdb
	}
	logger.Errorf("could not find cmdb %s", defaultCmdb)
	return nil
}

func RegisterCmdb(name string, cmdb ICmdb) {
	registryCmdb[name] = cmdb
}
