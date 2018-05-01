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
