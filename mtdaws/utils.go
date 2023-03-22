package mtdaws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/thefeli73/polemos/state"
)

// AwsInstance is basic info about a single aws instance (instance id, redion, pubIP and privIP)
type AwsInstance struct {
	InstanceID 		string
	Region			string
	PublicIP		string
	PrivateIP		string
}

// NewConfig creates a AWS config for a specific region
func NewConfig(region string, credentials string) aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigFiles([]string{credentials}), config.WithRegion(region))
	if err != nil {
		fmt.Println("Error creating config:", err)
		fmt.Println("Configure Credentials in line with the documentation found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials")
		os.Exit(1)
	}
	return cfg
}

// GetCloudID returns a string to find the instance based on information from aws
func GetCloudID(instance AwsInstance) string {
	return "aws_" + instance.Region + "_" + instance.InstanceID
}

// GetInstances scans all configured regions for instances and add them to services
func GetInstances(config state.Config) []AwsInstance {
	awsInstances := []AwsInstance{}
	for _, region := range config.AWS.Regions {
		awsConfig := NewConfig(region, config.AWS.CredentialsPath)
		instances, err := Instances(awsConfig)
		if err != nil {
			fmt.Println("Error listing instances:", err)
			continue
		}

		//fmt.Println("Listing instances in region:", region)
		for _, instance := range instances {
			var publicAddr string
			if instance.PublicIpAddress != nil {
				publicAddr = aws.ToString(instance.PublicIpAddress)
			}
			awsInstances = append(awsInstances, AwsInstance{
				InstanceID: aws.ToString(instance.InstanceId),
				Region: region,
				PublicIP: publicAddr, 
				PrivateIP: aws.ToString(instance.PrivateIpAddress)})
		}
	}
	return awsInstances
}

// PrintInstanceInfo prints info about a specific instance in a region
func PrintInstanceInfo(instance *types.Instance) {
	fmt.Println("\tInstance ID:", aws.ToString(instance.InstanceId))
	fmt.Println("\t\tInstance Type:", string(instance.InstanceType))
	fmt.Println("\t\tAMI ID:", aws.ToString(instance.ImageId))
	fmt.Println("\t\tState:", string(instance.State.Name))
	fmt.Println("\t\tAvailability Zone:", aws.ToString(instance.Placement.AvailabilityZone))
	if instance.PublicIpAddress != nil {
		fmt.Println("\t\tPublic IP Address:", aws.ToString(instance.PublicIpAddress))
	}
	fmt.Println("\t\tPrivate IP Address:", aws.ToString(instance.PrivateIpAddress))
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