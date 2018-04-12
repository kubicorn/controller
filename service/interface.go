package service

import (
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

type MachineMutation interface {
	Create(machine *clusterv1.Machine) (string, error)
	Exists(id string) bool
	Destroy(id string) error
	ListIDs(name string) ([]string, error)
}
