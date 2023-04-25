package mtdaws

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/uuid"
	"github.com/thefeli73/polemos/pcsdk"
	"github.com/thefeli73/polemos/state"
)

// AWSMoveInstance moves a specified instance to a new availability region
func AWSMoveInstance(config state.Config) (state.Config) {

	// pseudorandom instance from all services for testing
	var serviceUUID state.CustomUUID
	var instance state.Service
	for key, service := range config.MTD.Services {
		serviceUUID = key
		instance = service
		if !instance.AdminEnabled {continue}
		if !instance.Active {continue}
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

	//Create image
	t := time.Now()
	imageName, err := createImage(svc, instanceID)
	if err != nil {
		fmt.Println("Error creating image:\t", err)
		return config
	}
	fmt.Printf("Created image:\t\t%s (took %s)\n", imageName, time.Since(t).Round(100*time.Millisecond).String())

	// Wait for image
	t = time.Now()
	err = waitForImageReady(svc, imageName, 5*time.Minute)
	if err != nil {
		fmt.Println("Error waiting for image to be ready:\t", err)
		return config
	}
	fmt.Printf("Image is ready:\t\t%s (took %s)\n", imageName, time.Since(t).Round(100*time.Millisecond).String())

	// Launch new instance
	t = time.Now()
	newInstanceID, err := launchInstance(svc, realInstance, imageName, region)
	if err != nil {
		fmt.Println("Error launching instance:\t", err)
		return config
	}
	fmt.Printf("Launched new instance:\t%s (took %s)\n", newInstanceID, time.Since(t).Round(100*time.Millisecond).String())

	// Wait for instance
	t = time.Now()
	err = waitForInstanceReady(svc, newInstanceID, 5*time.Minute)
	if err != nil {
		fmt.Println("Error waiting for instance to be ready:\t", err)
		return config
	}
	fmt.Printf("instance is ready:\t\t%s (took %s)\n", newInstanceID, time.Since(t).Round(100*time.Millisecond).String())

	// Reconfigure Proxy to new instance
	t = time.Now()
	m := pcsdk.NewCommandModify(instance.ServicePort, instance.ServiceIP, serviceUUID)
	err = m.Execute(netip.AddrPortFrom(instance.EntryIP, config.MTD.ManagementPort))
	if err != nil {
		fmt.Printf("error executing modify command: %s\n", err)
		return config
	}
	fmt.Printf("Proxy modified. (took %s)\n", time.Since(t).Round(100*time.Millisecond).String())
	
	
	AWSUpdateService(config, region, serviceUUID, newInstanceID)

	cleanupAWS(svc, config, instanceID, imageName)


	return config
}

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

// cleanupAWS terminates the old instance, deregisters the image and deletes the old snapshot
func cleanupAWS(svc *ec2.Client, config state.Config, instanceID string, imageName string) state.Config {
	// Terminate old instance
	t := time.Now()
	err := terminateInstance(svc, instanceID)
	if err != nil {
		fmt.Println("Error terminating instance:\t", err)
		return config
	}
	fmt.Printf("Killed old instance:\t%s (took %s)\n", instanceID, time.Since(t).Round(100*time.Millisecond).String())

	// Deregister old image
	t = time.Now()
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
	fmt.Printf("Deregistered image:\t%s (took %s)\n", imageName, time.Since(t).Round(100*time.Millisecond).String())

	// Delete old snapshot
	t = time.Now()
	if len(image.BlockDeviceMappings) > 0 {
		snapshotID := aws.ToString(image.BlockDeviceMappings[0].Ebs.SnapshotId)
		err = deleteSnapshot(svc, snapshotID)
		if err != nil {
			fmt.Println("Error deleting snapshot:\t", err)
			return config
		}
		fmt.Printf("Deleted snapshot:\t%s (took %s)\n", snapshotID, time.Since(t).Round(100*time.Millisecond).String())
	}
	return config
}
