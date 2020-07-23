package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/toolkits/pkg/logger"
)

type LoggerSection struct {
	Dir       string `yaml:"dir"`
	Level     string `yaml:"level"`
	KeepHours uint   `yaml:"keepHours"`
}

func Init(l LoggerSection) {

	lb, err := logger.NewFileBackend(l.Dir)
	if err != nil {
		fmt.Println("cannot init logger:", err)
		os.Exit(1)
	}

	lb.SetRotateByHour(true)
	lb.SetKeepHours(l.KeepHours)

	logger.SetLogging(l.Level, lb)
}

func TimeoutWarning(tag, detailed string, start time.Time, timeLimit float64) {
	dis := time.Now().Sub(start).Seconds()
	if dis > timeLimit {
		logger.Warning(tag, " detailed:", detailed, "TimeoutWarning using", dis, "s")
	} else {
		logger.Info(tag, " detailed:", detailed, "Execution time", dis, "s")
	}
}
