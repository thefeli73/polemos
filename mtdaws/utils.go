package mtdaws

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

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

// DecodeCloudID returns information to locate instance in aws
func DecodeCloudID(cloudID string) (string, string) {
	split := strings.Split(cloudID, "_")
	if len(split) != 3 {
		panic(cloudID + " does not decode as AWS CloudID")
	}
	region := split[1]
	instanceID := split[2]
	return region, instanceID
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

// Instances returns all instances for a config i.e. a region
func Instances(config aws.Config) ([]types.Instance, error) {
	svc := ec2.NewFromConfig(config)

	input := &ec2.DescribeInstancesInput{}
	var instances []types.Instance

	paginator := ec2.NewDescribeInstancesPaginator(svc, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				instances = append(instances, instance)
			}
		}
	}
	return instances, nil
}

// createImage will create an AMI (amazon machine image) of a given instance
func createImage(svc *ec2.Client, instanceID string) (string, error) {
	input := &ec2.CreateImageInput{
		InstanceId:  aws.String(instanceID),
		Name:        aws.String(fmt.Sprintf("backup-%s-%d", instanceID, time.Now().Unix())),
		Description: aws.String("Migration backup"),
		NoReboot:    aws.Bool(true),
	}

	output, err := svc.CreateImage(context.TODO(), input)
	if err != nil {
		return "", err
	}

	return aws.ToString(output.ImageId), nil
}

// waitForImageReady polls every second to see if the image is ready
func waitForImageReady(svc *ec2.Client, imageID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return errors.New("timed out waiting for image to be ready")
		case <-time.After(1 * time.Second):
			input := &ec2.DescribeImagesInput{
				ImageIds: []string{imageID},
			}
			output, err := svc.DescribeImages(ctx, input)
			if err != nil {
				return err
			}

			if len(output.Images) > 0 && output.Images[0].State == types.ImageStateAvailable {
				return nil
			}
		}
	}
}

// waitForInstanceReady waits for the newly launched instance to be running and ready 
func waitForInstanceReady(svc *ec2.Client, newInstanceID string, timeout time.Duration) error {
	// Wait for the instance to be running
	waitInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{newInstanceID},
	}
	waiter := ec2.NewInstanceRunningWaiter(svc)
	err := waiter.Wait(context.TODO(), waitInput, timeout)
	if err != nil {
		return err
	}
	return nil
}

// launchInstance launches a instance IN RANDOM AVAILABILITY ZONE within the same region, based on an oldInstance and AMI (duplicating the instance)
func launchInstance(svc *ec2.Client, oldInstance *types.Instance, imageID string, region string) (string, error) {
	securityGroupIds := make([]string, len(oldInstance.SecurityGroups))
	for i, sg := range oldInstance.SecurityGroups {
		securityGroupIds[i] = aws.ToString(sg.GroupId)
	}
	availabilityZone, err := getRandomDifferentAvailabilityZone(svc, oldInstance, region)
	if err != nil {
		return "", err
	}
	var nameTag string
	for _, tag := range oldInstance.Tags {
		if aws.ToString(tag.Key) == "Name" {
			nameTag = aws.ToString(tag.Value)
			break
		}
	}

	input := &ec2.RunInstancesInput{
		ImageId:         aws.String(imageID),
		InstanceType:    oldInstance.InstanceType,
		MinCount:        aws.Int32(1),
		MaxCount:        aws.Int32(1),
		KeyName:         oldInstance.KeyName,
		SecurityGroupIds: securityGroupIds,
		Placement: &types.Placement{
			AvailabilityZone: aws.String(availabilityZone),
		},
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(nameTag),
					},
				},
			},
		},
	}

	output, err := svc.RunInstances(context.TODO(), input)
	if err != nil {
		return "", err
	}

	// TODO: save/index config for the new instance

	return aws.ToString(output.Instances[0].InstanceId), nil
}

// getRandomDifferentAvailabilityZone fetches all AZ from the same region as the instance and returns a random AZ that is not equal to the one used by the instance
func getRandomDifferentAvailabilityZone(svc *ec2.Client, instance *types.Instance, region string) (string, error) {
	// Seed the random generator
	rand.Seed(time.Now().UnixNano())

	// Get the current availability zone of the instance
	currentAZ := aws.ToString(instance.Placement.AvailabilityZone)

	// Describe availability zones in the region
	input := &ec2.DescribeAvailabilityZonesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("region-name"),
				Values: []string{region},
			},
		},
	}

	output, err := svc.DescribeAvailabilityZones(context.TODO(), input)
	if err != nil {
		return "", err
	}

	// Filter out the current availability zone
	availableAZs := []string{}
	for _, az := range output.AvailabilityZones {
		if aws.ToString(az.ZoneName) != currentAZ {
			availableAZs = append(availableAZs, aws.ToString(az.ZoneName))
		}
	}

	// If no other availability zones are available, return an error
	if len(availableAZs) == 0 {
		return "", errors.New("no other availability zones available")
	}

	// Select a random availability zone from the remaining ones
	randomIndex := rand.Intn(len(availableAZs))
	randomAZ := availableAZs[randomIndex]
	return randomAZ, nil
}


// terminateInstance kills an instance by id
func terminateInstance(svc *ec2.Client, instanceID string) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}
	_, err := svc.TerminateInstances(context.TODO(), input)
	return err
}

// describeImage gets info about an image from string
func describeImage(svc *ec2.Client, imageID string) (*types.Image, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: []string{imageID},
	}

	output, err := svc.DescribeImages(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	if len(output.Images) == 0 {
		return nil, errors.New("image not found")
	}

	return &output.Images[0], nil
}

// deregisterImage deletes the AMI passed as string
func deregisterImage(svc *ec2.Client, imageID string) error {
	input := &ec2.DeregisterImageInput{
		ImageId: aws.String(imageID),
	}

	_, err := svc.DeregisterImage(context.TODO(), input)
	return err
}

// deleteSnapshot deletes the snapshot passed as string
func deleteSnapshot(svc *ec2.Client, snapshotID string) error {
	input := &ec2.DeleteSnapshotInput{
		SnapshotId: aws.String(snapshotID),
	}

	_, err := svc.DeleteSnapshot(context.TODO(), input)
	return err
}

// getInstanceDetailsFromString does what the name says
func getInstanceDetailsFromString(svc *ec2.Client, instanceID string) (*types.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	output, err := svc.DescribeInstances(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &output.Reservations[0].Instances[0], nil
}
