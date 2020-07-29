package meicai

import (
	"fmt"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
	"github.com/toolkits/pkg/str"
)

func (m *Meicai) EndpointGet(col string, val interface{}) (*dataobj.Endpoint, error) {
	var obj dataobj.Endpoint

	return &obj, nil
}

func (m *Meicai) EndpointGets(query, batch, field string, limit, offset int) ([]dataobj.Endpoint, int64, error) {
	var objs []dataobj.Endpoint
	return objs, 0, nil
}

func (m *Meicai) EndpointUnderNodeGets(leafids []int64, query, batch, field string, limit,
	offset int) ([]dataobj.Endpoint, int64, error) {
	var objs []dataobj.Endpoint

	return objs, 0, nil
}

func (m *Meicai) EndpointIdsByIdents(idents []string) ([]int64, error) {
	idents = str.TrimStringSlice(idents)
	if len(idents) == 0 {
		return []int64{}, nil
	}

	ret := make([]int64, 0)

	return ret, nil
}

func (m *Meicai) EndpointBindings(endpointIds []int64) ([]dataobj.EndpointBinding, error) {

	ret := make([]dataobj.EndpointBinding, 0)
	return ret, nil
}

func (m *Meicai) EndpointUnderLeafs(leafIds []int64) ([]dataobj.Endpoint, error) {
	var endpoints []dataobj.Endpoint
	if len(leafIds) == 0 {
		return []dataobj.Endpoint{}, nil
	}

	return endpoints, nil
}

func (m *Meicai) Update(e *dataobj.Endpoint, cols ...string) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "Updata")
}

func (m *Meicai) EndpointImport(endpoints []string) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "EndpointImport")
}

func (m *Meicai) EndpointDel(ids []int64) error {
	return fmt.Errorf("meicai cmdb not impliement %s interface", "EndpointDel")
}
