package mtdaws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/uuid"
	"github.com/thefeli73/polemos/state"
)

// NewConfig creates a AWS config for a specific region
func NewConfig(region string, credentials string) aws.Config {
	// Create a new AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigFiles([]string{credentials}), config.WithRegion(region))
	if err != nil {
		fmt.Println("Error creating config:", err)
		fmt.Println("Configure Credentials in line with the documentation found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials")
		os.Exit(1)
	}
	return cfg
}

// IndexInstances scans all configured regions for instances and add them to services
func IndexInstances(config state.Config) {
	for _, region := range config.AWS.Regions {
		//fmt.Println("Listing instances in region:", region)
		awsConfig := NewConfig(region, config.AWS.CredentialsPath)
		instances, err := Instances(awsConfig)
		if err != nil {
			fmt.Println("Error listing instances:", err)
			continue
		}

		for _, instance := range instances {
			InstanceInfo(awsConfig, *instance.InstanceId)
			u := uuid.New()
			_ = u
		}
	}
}

// InstanceInfo collects info about a specific instance in a region
func InstanceInfo(config aws.Config, instanceID string) {
	// Create a new EC2 service client
	svc := ec2.NewFromConfig(config)

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	result, err := svc.DescribeInstances(context.TODO(), input)
	if err != nil {
		fmt.Println("Error describing instance:", err)
		return
	}
	// Print instance information
	instance := result.Reservations[0].Instances[0]	
	fmt.Println("Instance ID:", aws.ToString(instance.InstanceId))
	fmt.Println("Instance Type:", string(instance.InstanceType))
	fmt.Println("AMI ID:", aws.ToString(instance.ImageId))
	fmt.Println("State:", string(instance.State.Name))
	fmt.Println("Availability Zone:", aws.ToString(instance.Placement.AvailabilityZone))
	if instance.PublicIpAddress != nil {
		fmt.Println("Public IP Address:", aws.ToString(instance.PublicIpAddress))
	}
	fmt.Println("Private IP Address:", aws.ToString(instance.PrivateIpAddress))
}

// Instances returns all instances for a config i.e. a region
func Instances(config aws.Config) ([]*types.Instance, error) {
	svc := ec2.NewFromConfig(config)

	input := &ec2.DescribeInstancesInput{}
	var instances []*types.Instance

	paginator := ec2.NewDescribeInstancesPaginator(svc, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				instances = append(instances, &instance)
			}
		}
	}

	return instances, nil
}