package aws

import (
	"github.com/kubicorn/controller/machine"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

type AWSMachine struct {
}

func New() machine.MachineMutation {
	return &AWSMachine{}
}

func (a *AWSMachine) Create(machine *clusterv1.Machine) error {
	return nil
}

func (a *AWSMachine) Get(name string) (*clusterv1.Machine, error) {
	return &clusterv1.Machine{}, nil
}

func (a *AWSMachine) Destroy(machine *clusterv1.Machine) {
	return
}
