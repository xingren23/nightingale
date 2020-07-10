package backend

import (
	"github.com/didi/nightingale/src/modules/transfer/backend/influxdb"
	"github.com/didi/nightingale/src/modules/transfer/backend/tsdb"
)

type BackendSection struct {
	DataSource string `yaml:"datasource"`
	StraPath   string `yaml:"straPath"`

	Judge    JudgeSection             `yaml:"judge"`
	Tsdb     tsdb.TsdbSection         `yaml:"tsdb"`
	Influxdb influxdb.InfluxdbSection `yaml:"influxdb"`
	OpenTsdb OpenTsdbSection          `yaml:"opentsdb"`
	Kafka    KafkaSection             `yaml:"kafka"`
}

var (
	defaultDataSource    string
	StraPath             string
	tsdbDataSource       *tsdb.TsdbDataSource
	openTSDBPushEndpoint *OpenTsdbPushEndpoint
	influxdbDataSource   *influxdb.InfluxdbDataSource
	kafkaPushEndpoint    *KafkaPushEndpoint
)

func Init(cfg BackendSection) {
	defaultDataSource = cfg.DataSource
	StraPath = cfg.StraPath

	// init judge
	InitJudge(cfg.Judge)

	// init tsdb storage
	if cfg.Tsdb.Enabled {
		tsdbDataSource = &tsdb.TsdbDataSource{
			Section:               cfg.Tsdb,
			SendQueueMaxSize:      DefaultSendQueueMaxSize,
			SendTaskSleepInterval: DefaultSendTaskSleepInterval,
		}
		tsdbDataSource.Init()
		RegisterDataSource(tsdbDataSource.Section.Name, tsdbDataSource)
	}

	// init influxdb storage
	if cfg.Influxdb.Enabled {
		influxdbDataSource = &influxdb.InfluxdbDataSource{
			Section:               cfg.Influxdb,
			SendQueueMaxSize:      DefaultSendQueueMaxSize,
			SendTaskSleepInterval: DefaultSendTaskSleepInterval,
		}
		influxdbDataSource.Init()
		RegisterDataSource(influxdbDataSource.Section.Name, influxdbDataSource)

	}
	// init opentsdb storage
	if cfg.OpenTsdb.Enabled {
		openTSDBPushEndpoint = &OpenTsdbPushEndpoint{
			Section: cfg.OpenTsdb,
		}
		openTSDBPushEndpoint.Init()
	}
	// init kafka
	if cfg.Kafka.Enabled {
		kafkaPushEndpoint = &KafkaPushEndpoint{
			Section: cfg.Kafka,
		}
		kafkaPushEndpoint.Init()
	}
}
