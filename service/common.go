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

package service

import (
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-deploy/cluster-api/client"
	"k8s.io/kube-deploy/cluster-api/util"

	"fmt"

	"encoding/json"

	"strings"

	"github.com/kubicorn/controller/backoff"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

func RunService(cfg *ServiceConfiguration) {

	logger.Info("Starting controller loop...")
	errchan := ConcurrentReconcileMachines(cfg)
	for {
		select {
		case e1 := <-errchan:
			logger.Warning(e1.Error())
		}
	}
}

func ConcurrentReconcileMachines(cfg *ServiceConfiguration) chan error {
	ch := make(chan error)
	mm := cfg.CloudProvider
	t := backoff.NewBackoff("crm")
	go func() {
		for {
			t.Hang()
			cm, err := getClientMeta(cfg)
			if err != nil {
				ch <- fmt.Errorf("Unable to authenticate client: %v", err)
				continue
			}
			listOptions := metav1.ListOptions{}
			machines, err := cm.client.Machines().List(listOptions)
			if err != nil {
				ch <- fmt.Errorf("Unable to list machines: %v", err)
				continue
			}
			var name string
			for _, machine := range machines.Items {
				//pc := getProviderConfig(machine.Spec.ProviderConfig)
				exists := mm.Exists(machine.Name)
				if !exists {
					// Machine does not exist, create it
					id, err := mm.Create(&machine)
					if err != nil {
						ch <- fmt.Errorf("Unable to create machine [%s]: %v", machine.Name, err)
						continue
					}
					logger.Info("New machine id: %s", id)
					pc := getProviderConfig(machine.Spec.ProviderConfig)
					//
					// Here we can update the ProviderConfig
					// pc.ServerPool.Name = id

					pcBytes, err := json.Marshal(pc)
					if err != nil {
						ch <- fmt.Errorf("Unable to marshal new provider config! %v", err)
						continue
					}
					pcStr := string(pcBytes)
					machine.Spec.ProviderConfig = pcStr
					cm.client.Machines().Update(&machine)
					logger.Debug("Created machine: %s", machine.Name)
					continue
				}
				logger.Debug("Machine already exists: %s", machine.Name)
				spl := strings.Split(machine.Name, ".")
				name = spl[0]
			}
			names, err := mm.ListIDs(name)
			logger.Always("%+v", names)
			if err != nil {
				ch <- fmt.Errorf("Unable to list IDs: %v", err)
				continue
			}
			for _, n := range names {
				found := false
				for _, machine := range machines.Items {
					//pc := getProviderConfig(machine.Spec.ProviderConfig)
					if machine.Name == n {
						found = true
					}
				}
				if !found {
					mm.Destroy(n)
				}
			}

		}
	}()
	return ch
}

type crdClientMeta struct {
	client    *client.ClusterAPIV1Alpha1Client
	clientset *apiextensionsclient.Clientset
}

func getClientMeta(cfg *ServiceConfiguration) (*crdClientMeta, error) {
	kubeConfigPath, err := cfg.GetFilePath()
	if err != nil {
		return nil, err
	}
	client, err := util.NewApiClient(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	cs, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	clientMeta := &crdClientMeta{
		client:    client,
		clientset: cs,
	}
	return clientMeta, nil
}

func getProviderConfig(providerConfig string) *cluster.MachineProviderConfig {
	//logger.Info(providerConfig)
	mp := cluster.MachineProviderConfig{
		ServerPool: &cluster.ServerPool{},
	}
	json.Unmarshal([]byte(providerConfig), &mp)
	return &mp
}
