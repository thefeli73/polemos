package mtd_aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)
func New_config(region string) aws.Config {
	// Create a new AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		fmt.Println("Error creating config:", err)
		fmt.Println("Configure Credentials in line with the documentation found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials")
		os.Exit(1)
	}
	return cfg
}
func Instance_info(config aws.Config, instanceID string) {
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