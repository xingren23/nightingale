package n9e

import (
	"fmt"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
)

func (c *N9e) AppInstanceGets(string, string, string, int, int) ([]dataobj.AppInstance, int64, error) {
	return []dataobj.AppInstance{}, 0, fmt.Errorf("n9e cmdb not impliement %s interface", "AppInstanceGets")
}
