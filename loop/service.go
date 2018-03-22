package loop

import "github.com/kubicorn/kubicorn/pkg/logger"

func RunService() {

	logger.Info("Starting infinite loop...")
	for {}
}
