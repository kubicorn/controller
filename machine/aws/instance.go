// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
