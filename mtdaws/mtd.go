package mtdaws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/thefeli73/polemos/state"
)

// AWSMoveInstance moves a specified instance to a new availability region
func AWSMoveInstance(config state.Config) (state.Config) {
	instance := config.MTD.Services[0] // for testing use first instance
	region, instanceID := DecodeCloudID(instance.CloudID)
	awsConfig := NewConfig(region, config.AWS.CredentialsPath)
	svc := ec2.NewFromConfig(awsConfig)

	realInstance, err := getInstanceDetailsFromString(svc, instanceID)
	if err != nil {
		fmt.Println("Error getting instance details:\t", err)
		return config
	}

	imageName, err := createImage(svc, instanceID)
	if err != nil {
		fmt.Println("Error creating image:\t", err)
		return config
	}
	fmt.Println("Created image:\t", imageName)

	err = waitForImageReady(svc, imageName, 5*time.Minute)
	if err != nil {
		fmt.Println("Error waiting for image to be ready:\t", err)
		return config
	}
	fmt.Println("Image is ready:\t", imageName)

	newInstanceID, err := launchInstance(svc, realInstance, imageName)
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
	fmt.Println("Terminated old instance:\t", instanceID)

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

	return config
}
