package loop

import (
	"github.com/kubicorn/kubicorn/pkg/logger"
	"fmt"
)

type ServiceOptions struct {

}

type Service struct {

}

func InitializeService(options *ServiceOptions) (*Service, error) {
	service := &Service{}
	return service, nil
}

func RunService(options *ServiceOptions) error {
	svc, err := InitializeService(options)
	if err != nil {
		return fmt.Errorf("Unable to initialize service: %v", err)
	}
	logger.Info("Starting control loop...")
	for {
		safeState := AtomicGetState()
		safeState.AtomicEnsureAttempt(svc)
	}
}
