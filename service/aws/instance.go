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
	"fmt"
	"time"

	"encoding/json"
	"sync"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kubicorn/controller/service"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/logger"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

type AWSMachine struct {
	EC2 *ec2.EC2
}

var (
	infrastructureMutex            sync.Mutex
	checkForExistsAfterCreateCount = 225
)

func New(region string, profile string) (service.MachineMutation, error) {

	session, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(region)},
		// Support MFA when authing using assumed roles.
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		Profile:                 profile,
	})
	if err != nil {
		return nil, err
	}
	ec2Session := ec2.New(session)
	return &AWSMachine{
		EC2: ec2Session,
	}, nil
}

func (a *AWSMachine) Create(machine *clusterv1.Machine) (string, error) {
	infrastructureMutex.Lock()
	defer infrastructureMutex.Unlock()
	pc := getProviderConfig(machine.Spec.ProviderConfig)
	spl := strings.Split(machine.Name, ".")
	clusterName := spl[0]
	if pc.ServerPool.Image == "" {
		return "", nil
	}
	logger.Debug("AMI: %s
", pc.ServerPool.Image)
	logger.Debug("KeyPair: %s
", clusterName)

	// Calculate Security Groups
	var sgs []*string
	for _, firewall := range pc.ServerPool.Firewalls {
		sgs = append(sgs, &firewall.Identifier)
	}

	// Calculate Subnet IDs and Create NIC
	var nics []*ec2.InstanceNetworkInterfaceSpecification
	index := 0
	for _, subnet := range pc.ServerPool.Subnets {
		nic := &ec2.InstanceNetworkInterfaceSpecification{
			SubnetId:                 aws.String(subnet.Identifier),
			AssociatePublicIpAddress: aws.Bool(true),
			DeviceIndex:              aws.Int64(int64(index)),
			Groups:                   sgs,
		}
		index++
		nics = append(nics, nic)
	}

	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(pc.ServerPool.Image),
		InstanceType: aws.String(pc.ServerPool.Size),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      aws.String(clusterName),
		UserData:     aws.String(string(pc.ServerPool.GeneratedNodeUserData)),
		//SecurityGroups:    sgs,
		NetworkInterfaces: nics,
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(pc.ServerPool.InstanceProfile.Name),
		},
	}
	output, err := a.EC2.RunInstances(input)
	if err != nil {
		return "", fmt.Errorf("Unable to create instance: %v", err)
	}
	tagInput := &ec2.CreateTagsInput{
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("KubernetesCluster"),
				Value: aws.String(clusterName),
			},
			{
				Key:   aws.String("Name"),
				Value: aws.String(machine.Name),
			},
		},
		Resources: []*string{
			output.Instances[0].InstanceId,
		},
	}
	_, err = a.EC2.CreateTags(tagInput)
	if err != nil {
		defer a.Destroy(machine.Name)
		logger.Warning("Unable to tag instance: destroying: %v", err)
		return "", fmt.Errorf("Unable to tag instance: %v", err)
	}

	// --- Hang until instance registers
	i := 0
	for {
		if a.Exists(machine.Name) {
			return *output.Instances[0].InstanceId, nil
		}
		time.Sleep(time.Second * 5)
		i++
		if i == checkForExistsAfterCreateCount {
			return "", fmt.Errorf("Unable to detect instance after create and minimual checks expired")
		}
		logger.Info("Waiting for machine to register [%s]...", machine.Name)
	}
	logger.Always("Created instance: %s", *output.Instances[0].InstanceId)
	return *output.Instances[0].InstanceId, nil
}

func (a *AWSMachine) Exists(name string) bool {
	if name == "" {
		logger.Info("Empty name")
		return true
	}
	logger.Info("Query for instance name [%s]", name)
	//infrastructureMutex.Lock()
	//defer infrastructureMutex.Unlock()
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}
	output, err := a.EC2.DescribeInstances(input)
	if err != nil {
		logger.Warning("Unable to describe instances: %v", err)
		// If there is an error we want to assume it exists so we don't take action
		return true
	}
	if len(output.Reservations) < 1 {
		logger.Always("Machine [%s] NOT found", name)
		return false
	}

	logger.Debug("Reservation count: %d", len(output.Reservations))
	for _, reservation := range output.Reservations {
		instances := reservation.Instances
		logger.Debug("Instance count: %d", len(instances))
		for _, instance := range instances {
			if *instance.State.Name != "running" {
				logger.Debug("Instance not `running` state is [%s]", *instance.State.Name)
				continue
			}
			tags := instance.Tags
			for _, tag := range tags {
				k := *tag.Key
				v := *tag.Value
				if k == "Name" {
					if v == name {
						logger.Always("Machine [%s] found", name)
						return true
					}
					logger.Debug("Instance found but not matched [%s][%s]", name, v)
				}
			}
		}
	}

	logger.Always("Machine [%s] NOT found", name)
	return false
}

func (a *AWSMachine) Destroy(name string) error {
	logger.Always("Destroying instance: %s", name)
	infrastructureMutex.Lock()
	defer infrastructureMutex.Unlock()
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}
	output, err := a.EC2.DescribeInstances(input)
	if err != nil {
		return fmt.Errorf("Unable to destroy instance: %v", err)
	}
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Name == "running" {
				input := &ec2.TerminateInstancesInput{
					InstanceIds: []*string{
						instance.InstanceId,
					},
				}
				_, err := a.EC2.TerminateInstances(input)
				if err != nil {
					return fmt.Errorf("Unable to destroy instance [%s]: %v", name, err)
				}
				logger.Always("Terminated instance: %s", name)
			}
		}
	}

	return nil
}

func (a *AWSMachine) ListIDs(name string) ([]string, error) {
	logger.Always("List IDs for cluster [%s]", name)
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:KubernetesCluster"),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}
	output, err := a.EC2.DescribeInstances(input)
	if err != nil {
		return []string{}, fmt.Errorf("Unable to list ids: %v", err)
	}
	var names []string
	if len(output.Reservations) == 0 {
		return names, nil
	}
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Name != "running" {
				continue
			}
			tags := instance.Tags
			for _, tag := range tags {
				k := *tag.Key
				v := *tag.Value
				if k == "Name" {
					names = append(names, v)
				}
			}
		}
	}
	return names, nil
}

func getProviderConfig(providerConfig string) *cluster.MachineProviderConfig {
	//logger.Info(providerConfig)
	mp := cluster.MachineProviderConfig{
		ServerPool: &cluster.ServerPool{},
	}
	json.Unmarshal([]byte(providerConfig), &mp)
	return &mp
}
