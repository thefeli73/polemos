package mtdaws

import (
	"fmt"
	"time"

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
		fmt.Println("Error getting instance details:", err)
		return config
	}

	imageName, err := createImage(svc, instanceID)
	if err != nil {
		fmt.Println("Error creating image:", err)
		return config
	}
	fmt.Println("Created image: ", imageName)

	err = waitForImageReady(svc, imageName, 5*time.Minute)
	if err != nil {
		fmt.Println("Error waiting for image to be ready:", err)
		return config
	}
	fmt.Println("Image is ready:", imageName)

	newInstanceID, err := launchInstance(svc, realInstance, imageName)
	if err != nil {
		fmt.Println("Error launching instance:", err)
		return config
	}
	fmt.Println("Launched new instance:", newInstanceID)

	err = terminateInstance(svc, instanceID)
	if err != nil {
		fmt.Println("Error terminating instance:", err)
		return config
	}
	fmt.Println("Terminated original instance:", instanceID)


	return config
}
