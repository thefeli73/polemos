package mtdaws

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/uuid"
	"github.com/thefeli73/polemos/state"
)

// AWSUpdateService updates a specified service config to match a newly moved instance
func AWSUpdateService(config state.Config, region string, service state.CustomUUID, newInstanceID string) (state.Config) {
	awsConfig := NewConfig(region, config.AWS.CredentialsPath)
	svc := ec2.NewFromConfig(awsConfig)
	instance, err := getInstanceDetailsFromString(svc, newInstanceID)
	if err != nil {
		fmt.Println("Error getting instance details:\t", err)
		return config
	}

	var publicAddr string
	if instance.PublicIpAddress != nil {
		publicAddr = aws.ToString(instance.PublicIpAddress)
	}
	formattedinstance := AwsInstance{
		InstanceID: aws.ToString(instance.InstanceId),
		Region: region,
		PublicIP: publicAddr, 
		PrivateIP: aws.ToString(instance.PrivateIpAddress),
	}
	cloudid := GetCloudID(formattedinstance)
	serviceip := netip.MustParseAddr(publicAddr)
	config.MTD.Services[service] = state.Service{CloudID: cloudid, ServiceIP: serviceip}
	return config
}


// isInstanceRunning returns if an instance is running (true=running)
func isInstanceRunning(instance *types.Instance) bool {
	return instance.State.Name == types.InstanceStateNameRunning
}

// AWSMoveInstance moves a specified instance to a new availability region
func AWSMoveInstance(config state.Config) (state.Config) {

	// pseudorandom instance from all services for testing
	var serviceUUID state.CustomUUID
	var instance state.Service
	for key, service := range config.MTD.Services {
		serviceUUID = key
		instance = service
		break
	}

	fmt.Println("MTD move service:\t", uuid.UUID.String(uuid.UUID(serviceUUID)))

	region, instanceID := DecodeCloudID(instance.CloudID)
	awsConfig := NewConfig(region, config.AWS.CredentialsPath)
	svc := ec2.NewFromConfig(awsConfig)

	realInstance, err := getInstanceDetailsFromString(svc, instanceID)
	if err != nil {
		fmt.Println("Error getting instance details:\t", err)
		return config
	}

	if !isInstanceRunning(realInstance) {
		fmt.Println("Error, Instance is not running!")
		return config
	}
	if instance.AdminDisabled {
		fmt.Println("Error, Service is Disabled!")
		return config
	}
	if instance.Inactive {
		fmt.Println("Error, Service is Inactive!")
		return config
	}

	imageName, err := createImage(svc, instanceID)
	if err != nil {
		fmt.Println("Error creating image:\t", err)
		return config
	}
	fmt.Println("Created image:\t\t", imageName)

	err = waitForImageReady(svc, imageName, 5*time.Minute)
	if err != nil {
		fmt.Println("Error waiting for image to be ready:\t", err)
		return config
	}
	fmt.Println("Image is ready:\t\t", imageName)

	newInstanceID, err := launchInstance(svc, realInstance, imageName, region)
	if err != nil {
		fmt.Println("Error launching instance:\t", err)
		return config
	}
	fmt.Println("Launched new instance:\t", newInstanceID)

	err = terminateInstance(svc, instanceID)
	if err != nil {
		fmt.Println("Error terminating instance:\t", err)
		return config
	}
	fmt.Println("Killed old instance:\t", instanceID)

	image, err := describeImage(svc, imageName)
	if err != nil {
		fmt.Println("Error describing image:\t", err)
		return config
	}

	err = deregisterImage(svc, imageName)
	if err != nil {
		fmt.Println("Error deregistering image:\t", err)
		return config
	}
	fmt.Println("Deregistered image:\t", imageName)

	if len(image.BlockDeviceMappings) > 0 {
		snapshotID := aws.ToString(image.BlockDeviceMappings[0].Ebs.SnapshotId)
		err = deleteSnapshot(svc, snapshotID)
		if err != nil {
			fmt.Println("Error deleting snapshot:\t", err)
			return config
		}
		fmt.Println("Deleted snapshot:\t", snapshotID)
	}

	AWSUpdateService(config, region, serviceUUID, newInstanceID)

	return config
}
