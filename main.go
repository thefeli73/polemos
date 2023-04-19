package main

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/mtdaws"
	"github.com/thefeli73/polemos/state"
)

// ConfigPath is a string of the location for the configfile
var ConfigPath string

func main() {
	fmt.Println("Starting Polemos")

	ConfigPath = "config.yaml"
	
    // Initialize the config.Services map
	var config state.Config
	config.MTD.Services = make(map[state.CustomUUID]state.Service)

	config = state.LoadConf(ConfigPath)
	state.SaveConf(ConfigPath, config)


	config = indexAllInstances(config)
	state.SaveConf(ConfigPath, config)

	// START DOING MTD
	mtdLoop(config)
}

func mtdLoop(config state.Config) {
	for true {
		//TODO: figure out migration (MTD)
		config = movingTargetDefense(config)
		state.SaveConf(ConfigPath, config)

		fmt.Println("Sleeping for 5 seconds")
		time.Sleep(5*time.Second)

		//TODO: proxy commands
	}
}

func movingTargetDefense(config state.Config) state.Config{

	mtdaws.AWSMoveInstance(config)
	return config
}

func indexAllInstances(config state.Config) state.Config {
	fmt.Println("Indexing instances")
	t := time.Now()

	for _, service := range config.MTD.Services {
		service.Active = false
	}

	//index AWS instances
	awsNewInstanceCounter := 0
	awsInactiveInstanceCounter := len(config.MTD.Services)
	awsInstanceCounter := 0
	awsInstances := mtdaws.GetInstances(config)
	for _, instance := range awsInstances {
		cloudID := mtdaws.GetCloudID(instance)
		ip, err := netip.ParseAddr(instance.PublicIP)
		if err != nil {
			fmt.Println("Error converting ip:\t", err)
			continue
		}
		var found bool
		config, found = indexInstance(config, cloudID, ip)
		if !found {
			awsNewInstanceCounter++
		} else {
			awsInactiveInstanceCounter--
		}
		awsInstanceCounter++
	}
	// TODO: Purge instances in config that are not found in the cloud
	fmt.Printf("Found %d active AWS instances (%d newly added, %d inactive) (took %s)\n",
		awsInstanceCounter, awsNewInstanceCounter, awsInactiveInstanceCounter, time.Since(t).Round(100*time.Millisecond).String())


	return config
}

func indexInstance(config state.Config, cloudID string, serviceIP netip.Addr) (state.Config, bool) {
	found := false
	var foundUUID state.CustomUUID
	for u, service := range config.MTD.Services {
		if service.CloudID == cloudID {
			found = true
			foundUUID = u
			break;
		}
	}

	if !found {
		fmt.Println("New instance found:\t", cloudID)
		u := uuid.New()
		config.MTD.Services[state.CustomUUID(u)] = state.Service{CloudID: cloudID, ServiceIP: serviceIP, Active: true, AdminEnabled: true}
		state.SaveConf(ConfigPath, config)
		
	} else {
		s := config.MTD.Services[foundUUID]
		s.Active = true
		config.MTD.Services[foundUUID] = s
		state.SaveConf(ConfigPath, config)
	}
	return config, found
}
