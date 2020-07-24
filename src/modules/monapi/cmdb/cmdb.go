package cmdb

import (
	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
	"github.com/toolkits/pkg/logger"
)

type IEndpointReadable interface {
	EndpointTotal(query, batch, field string) (int64, error)
	EndpointGet(col string, val interface{}) (*dataobj.Endpoint, error)
	EndpointGets(query, batch, field string, limit, offset int) ([]dataobj.Endpoint, error)
	EndpointIdsByIdents(idents []string) ([]int64, error)
	EndpointUnderNodeTotal(leafids []int64, query, batch, field string) (int64, error)
	EndpointUnderNodeGets(leafids []int64, query, batch, field string, limit, offset int) ([]dataobj.Endpoint, error)
	EndpointUnderLeafs(leafIds []int64) ([]dataobj.Endpoint, error)
}

type IEndpoint interface {
	IEndpointReadable

	Update(e *dataobj.Endpoint, cols ...string) error
	EndpointDel(ids []int64) error
	EndpointImport(endpoints []string) error
	EndpointBindings(endpointIds []int64) ([]dataobj.EndpointBinding, error)
}

type INodeReadable interface {
	NodeGets(where string, args ...interface{}) (nodes []dataobj.Node, err error)
	NodeGetsByPaths(paths []string) ([]dataobj.Node, error)
	NodeByIds(ids []int64) ([]dataobj.Node, error)
	LeafIds(n *dataobj.Node) ([]int64, error)
	NodeQueryPath(query string, limit int) (nodes []dataobj.Node, err error)
	TreeSearchByPath(query string) (nodes []dataobj.Node, err error)
	NodeGet(col string, val interface{}) (*dataobj.Node, error)
	NodesGetByIds(ids []int64) ([]dataobj.Node, error)
	Pids(n *dataobj.Node) ([]int64, error)
}

type INode interface {
	INodeReadable

	InitNode()
	CreateChild(n *dataobj.Node, name string, leaf int, note string) (int64, error)
	Bind(n *dataobj.Node, endpointIds []int64, delOld int) error
	Unbind(n *dataobj.Node, hostIds []int64) error
	Del(n *dataobj.Node) error
	Rename(n *dataobj.Node, name string) error
}

type INodeEndpointReadable interface {
	NodeIdsGetByEndpointId(endpointId int64) ([]int64, error)
	EndpointIdsByNodeIds(nodeIds []int64) ([]int64, error)
	EndpointBindingsForMail(endpoints []string) []string
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

