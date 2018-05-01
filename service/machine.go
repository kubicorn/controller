package service

import (
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-deploy/cluster-api/client"
	"k8s.io/kube-deploy/cluster-api/util"

	"fmt"

<<<<<<< HEAD:machine/machine.go
	"github.com/kubicorn/controller/backoff"
=======
	"encoding/json"

	"github.com/kubicorn/controller/backoff"
	"github.com/kubicorn/kubicorn/apis/cluster"
>>>>>>> Switching computers after work:service/machine.go
	"github.com/kubicorn/kubicorn/pkg/logger"
)

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
					pc := getProviderConfig(machine.Spec.ProviderConfig)
					pc.ServerPool.Name = id
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
			}
			ids, err := mm.ListIDs(name)
			if err != nil {
				ch <- fmt.Errorf("Unable to list IDs: %v", err)
				continue
			}
			for _, id := range ids {
				found := false
				for _, machine := range machines.Items {
					pc := getProviderConfig(machine.Spec.ProviderConfig)
					if pc.ServerPool.Name == id {
						found = true
					}
				}
				if !found {
					mm.Destroy(id)
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
