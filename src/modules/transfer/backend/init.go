package backend

import (
	"github.com/didi/nightingale/src/modules/transfer/backend/influxdb"
	"github.com/didi/nightingale/src/modules/transfer/backend/tsdb"
)

type BackendSection struct {
	Storage  string `yaml:"storage"`
	StraPath string `yaml:"straPath"`

	Judge    JudgeSection             `yaml:"judge"`
	Tsdb     tsdb.TsdbSection         `yaml:"tsdb"`
	Influxdb influxdb.InfluxdbSection `yaml:"influxdb"`
	OpenTsdb OpenTsdbSection          `yaml:"opentsdb"`
	Kafka    KafkaSection             `yaml:"kafka"`
}

var (
	defaultStorage    string
	StraPath          string
	tsdbStorage       *tsdb.TsdbDatasource
	openTSDBStorage   *OpenTsdbPushEndpoint
	influxdbStorage   *influxdb.InfluxdbDatasource
	kafkaPushEndpoint *KafkaPushEndpoint
)

func Init(cfg BackendSection) {
	defaultStorage = cfg.Storage
	StraPath = cfg.StraPath

	// init judge
	InitJudge(cfg.Judge)

	// init tsdb storage
	if cfg.Tsdb.Enabled {
		tsdbStorage = &tsdb.TsdbDatasource{
			Section:               cfg.Tsdb,
			SendQueueMaxSize:      DefaultSendQueueMaxSize,
			SendTaskSleepInterval: DefaultSendTaskSleepInterval,
		}
		tsdbStorage.Init()
		RegisterStorage(tsdbStorage.Section.Name, tsdbStorage)
	}

	// init influxdb storage
	if cfg.Influxdb.Enabled {
		influxdbStorage = &influxdb.InfluxdbDatasource{
			Section:               cfg.Influxdb,
			SendQueueMaxSize:      DefaultSendQueueMaxSize,
			SendTaskSleepInterval: DefaultSendTaskSleepInterval,
		}
		influxdbStorage.Init()
		RegisterStorage(influxdbStorage.Section.Name, influxdbStorage)

	}
	// init opentsdb storage
	if cfg.OpenTsdb.Enabled {
		openTSDBStorage = &OpenTsdbPushEndpoint{
			Section: cfg.OpenTsdb,
		}
		openTSDBStorage.Init()
	}
	// init kafka
	if cfg.Kafka.Enabled {
		kafkaPushEndpoint = &KafkaPushEndpoint{
			Section: cfg.Kafka,
		}
		kafkaPushEndpoint.Init()
	}
}
