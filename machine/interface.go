package machine

import (
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

type MachineMutation interface {
	Create(machine *clusterv1.Machine) error
	Get(name string) (*clusterv1.Machine, error)
	Destroy(machine *clusterv1.Machine)
}
