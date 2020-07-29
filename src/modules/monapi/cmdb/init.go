package cmdb

import (
	"log"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/meicai"
	"github.com/didi/nightingale/src/modules/monapi/cmdb/n9e"
)

type CmdbSection struct {
	Default string               `yaml:"name""`
	N9e     n9e.N9eSection       `yaml:"n9e""`
	Meicai  meicai.MeicaiSection `yaml:"meicai"`
}

func Init(cmdb CmdbSection) {
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
			meicaiCmdb := &meicai.Meicai{}
			meicaiCmdb.Init()
			RegisterCmdb(cmdb.Meicai.Name, meicaiCmdb)
			defaultCmdb = cmdb.Default
		} else {
			log.Println("config invalid %s", cmdb.Default)
		}
	}
}
