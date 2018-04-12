package machine

import (
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-deploy/cluster-api/client"
	"k8s.io/kube-deploy/cluster-api/util"
)

func ConcurrentReconcileMachines(mm MachineMutation) chan error {
	ch := make(chan error)
	go func() {
		cm := getClientMeta()
	}()
	return ch
}

type crdClientMeta struct {
	client    *client.ClusterAPIV1Alpha1Client
	clientset *apiextensionsclient.Clientset
}

func getClientMeta() (*crdClientMeta, error) {
	kubeConfigPath := kubeconfig.GetKubeConfigPath(cluster)
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
