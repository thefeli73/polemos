package state

import (
	"fmt"
	"io/ioutil"
	"net/netip"
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Config contains all MTD services and cloud provider configs
type Config struct {
    MTD             mtdconf     `yaml:"mtd"`
    AWS             aws         `yaml:"aws"`
}

type mtdconf struct {
    Services        []Service   `yaml:"services"`
}

// Service contains all necessary information about a service to identify it in the cloud as well as configuring a proxy for it
type Service struct {
    ID              CustomUUID  `yaml:"id"`
    CloudID         string      `yaml:"cloud_id"`
    EntryIP         netip.Addr  `yaml:"entry_ip"`
    EntryPort       uint16      `yaml:"entry_port"`
    ServiceIP       netip.Addr  `yaml:"service_ip"`
    ServicePort     uint16      `yaml:"service_port"`
}

// CustomUUID is an alias for uuid.UUID to enable custom unmarshal function
type CustomUUID uuid.UUID

type aws struct {
    Regions         []string    `yaml:"regions"`
    CredentialsPath string      `yaml:"credentials_path"`
}

// UnmarshalYAML parses uuid in yaml to CustomUUID type
func (u *CustomUUID) UnmarshalYAML(value *yaml.Node) error {
	id, err := uuid.Parse(value.Value)
	if err != nil {
		return err
	}
	*u = CustomUUID(id)
	return nil
}

// MarshalYAML parses CustomUUID type to uuid string for yaml 
func (u CustomUUID) MarshalYAML() (interface{}, error) {
	return uuid.UUID(u).String(), nil
}

// LoadConf loads config from a yaml file
func LoadConf(filename string) (Config) {
    var config Config

    data, err := ioutil.ReadFile(filename)
    if err != nil {
        fmt.Println("Error reading file:", err)

		fmt.Println("Attempting to load default config")
        data, err = ioutil.ReadFile("config.default.yaml")
        if err != nil {
            fmt.Println("Error reading file:", err)
            os.Exit(1)
        }
    }

    err = yaml.Unmarshal([]byte(data), &config)
    if err != nil {
        fmt.Println("Error importing config:", err)
        os.Exit(1)
    }
    fmt.Println("Imported config succesfully!")
    return config
}

// SaveConf saves config to yaml file
func SaveConf(filename string, config Config) (error) {
    yamlBytes, err := yaml.Marshal(&config)
    if err != nil {
        return err
	}

	err = ioutil.WriteFile(filename, yamlBytes, 0644)
	if err != nil {
        return err
	}
    return nil
}