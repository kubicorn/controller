package aws

import (
	"fmt"

	"encoding/json"
	"sync"

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
	infrastructureMutex sync.Mutex
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
	userData := ""
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String("imageid"),
		InstanceType: aws.String("size"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      aws.String("ssh key"),
		UserData:     &userData,
	}
	output, err := a.EC2.RunInstances(input)
	if err != nil {
		return "", fmt.Errorf("Unable to create instance: %v", err)
	}
	tagInput := &ec2.CreateTagsInput{
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("KubicornCluster"),
				Value: aws.String(pc.Name),
			},
		},
		Resources: []*string{
			output.Instances[0].InstanceId,
		},
	}
	_, err = a.EC2.CreateTags(tagInput)
	if err != nil {
		defer a.Destroy(pc.Name)
		logger.Warning("Unable to tag instance: destroying: %v", err)
		return "", fmt.Errorf("Unable to tag instance: %v", err)
	}
	logger.Info("Created instance: %s", output.Instances[0].InstanceId)
	return *output.Instances[0].InstanceId, nil
}

func (a *AWSMachine) Exists(id string) bool {
	infrastructureMutex.Lock()
	defer infrastructureMutex.Unlock()
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			&id,
		},
	}
	output, err := a.EC2.DescribeInstances(input)
	if err != nil {
		logger.Warning("Unable to describe instances: %v", err)
		// If there is an error we want to assume it exists so we don't take action
		return true
	}
	if len(output.Reservations[0].Instances) > 1 {
		return true
	}
	return false
}

func (a *AWSMachine) Destroy(id string) error {
	infrastructureMutex.Lock()
	defer infrastructureMutex.Unlock()
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			&id,
		},
	}
	_, err := a.EC2.TerminateInstances(input)
	if err != nil {
		return fmt.Errorf("Unable to destroy instance [%s]: %v", id, err)
	}
	return nil
}

func (a *AWSMachine) ListIDs(name string) ([]string, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("KubicornCluster"),
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
	var ids []string
	for _, instance := range output.Reservations[0].Instances {
		ids = append(ids, *instance.InstanceId)
	}
	return ids, nil
}

func getProviderConfig(providerConfig string) *cluster.MachineProviderConfig {
	logger.Info(providerConfig)
	mp := cluster.MachineProviderConfig{
		ServerPool: &cluster.ServerPool{},
	}
	json.Unmarshal([]byte(providerConfig), &mp)
	return &mp
}
