package machine

import (
	"github.com/kubicorn/kubicorn/pkg/logger"
)

func RunService(cfg *ServiceConfiguration) {

	logger.Info("Starting infinite loop...")
	errchan := ConcurrentReconcileMachines(cfg)
	for {
		select {
		case e1 := <-errchan:
			logger.Warning(e1.Error())
		}
	}
}
