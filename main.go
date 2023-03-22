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

	config = indexInstances(config)

}

func indexInstances(config state.Config) state.Config {
	fmt.Println("Indexing instances")

	//index AWS instances
	awsInstances := mtdaws.GetInstances(config)
	for _, instance := range awsInstances {
		cloudID := mtdaws.GetCloudID(instance)
		ip, err := netip.ParseAddr(instance.PublicIP)
		if err != nil {
			fmt.Println("Error converting ip:", err)
			continue
		}
		config = indexInstance(config, cloudID, ip)
	}
	return config
}

func indexInstance(config state.Config, cloudID string, serviceIP netip.Addr) state.Config {
	for _, service := range config.MTD.Services {
		if service.CloudID == cloudID {
			return config
		}
	}
	u := uuid.New()
	newService := state.Service{
		ID: state.CustomUUID(u),
		CloudID: cloudID,
		ServiceIP: serviceIP}
	config.MTD.Services = append(config.MTD.Services, newService)
	state.SaveConf(ConfigPath, config)
	return config
}