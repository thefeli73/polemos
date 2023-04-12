package main

import (
	"fmt"
	"net/netip"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/mtdaws"
	"github.com/thefeli73/polemos/state"
)

// ConfigPath is a string of the location for the configfile
var ConfigPath string

func main() {
	fmt.Println("Starting Polemos")

	ConfigPath = "config.yaml"
	
	config := state.LoadConf(ConfigPath)
	state.SaveConf(ConfigPath, config)

	config = indexAllInstances(config)

	//TODO: figure out migration (MTD)
	config = movingTargetDefense(config)

	//TODO: proxy commands
}

func movingTargetDefense(config state.Config) state.Config{

	mtdaws.AWSMoveInstance(config)
	return config
}

func indexAllInstances(config state.Config) state.Config {
	fmt.Println("Indexing instances")

	//index AWS instances
	awsNewInstanceCounter := 0
	awsRemovedInstanceCounter := 0
	awsInstanceCounter := 0
	awsInstances := mtdaws.GetInstances(config)
	for _, instance := range awsInstances {
		cloudID := mtdaws.GetCloudID(instance)
		ip, err := netip.ParseAddr(instance.PublicIP)
		if err != nil {
			fmt.Println("Error converting ip:\t", err)
			continue
		}
		newService, found := indexInstance(config, cloudID, ip)
		if !found {
			fmt.Println("New instance found:\t", newService.CloudID)
			config.MTD.Services = append(config.MTD.Services, newService)
			state.SaveConf(ConfigPath, config)
			awsNewInstanceCounter++
		}
		awsInstanceCounter++
	}
	// TODO: Purge instances in config that are not found in the cloud
	fmt.Printf("Found %d AWS instances (%d newly added, %d removed)\n", awsInstanceCounter, awsNewInstanceCounter, awsRemovedInstanceCounter)


	return config
}

func indexInstance(config state.Config, cloudID string, serviceIP netip.Addr) (state.Service, bool) {
	found := false
	for _, service := range config.MTD.Services {
		if service.CloudID == cloudID {
			found = true
			break;
		}
	}
	u := uuid.New()
	newService := state.Service{
		ID: state.CustomUUID(u),
		CloudID: cloudID,
		ServiceIP: serviceIP}
	return newService, found
}
