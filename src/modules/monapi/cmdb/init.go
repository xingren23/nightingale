package cmdb

import (
	"log"

	"github.com/didi/nightingale/src/modules/monapi/config"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/meicai"
	"github.com/didi/nightingale/src/modules/monapi/cmdb/n9e"
)

func Init(cmdb config.CmdbSection) {
	log.Printf("init cmdb section %s", cmdb)

	if cmdb.N9e.Enabled {
		if cmdb.N9e.Name == cmdb.Default {
			log.Print("init n9e cmdb")
			n9eCmdb := &n9e.N9e{}
			n9eCmdb.Init()
			RegisterCmdb(cmdb.N9e.Name, n9eCmdb)
			defaultCmdb = cmdb.Default
		} else {
			log.Println("config invalid %s", cmdb.Default)
		}
	}

	if cmdb.Meicai.Enabled {
		if cmdb.Meicai.Name == cmdb.Default {
			log.Println("init meicai cmdb")
			meicaiCmdb := &meicai.Meicai{OpsAddr: cmdb.Meicai.OpsAddr, Timeout: cmdb.Meicai.Timeout}
			meicaiCmdb.Init()
			RegisterCmdb(cmdb.Meicai.Name, meicaiCmdb)
			defaultCmdb = cmdb.Default
		} else {
			log.Println("config invalid %s", cmdb.Default)
		}
	}
}
