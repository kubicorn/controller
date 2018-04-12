package service

import (
	"io/ioutil"
	"os"
)

type CloudProvider MachineMutation

type ServiceConfiguration struct {
	KubeConfigContent string
	cloudProviderName string
	CloudProvider     CloudProvider
}

func (s *ServiceConfiguration) GetFilePath() (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "kubicorn")
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(file.Name(), []byte(s.KubeConfigContent), 0755)
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}
