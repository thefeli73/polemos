package state

import (
	"io/ioutil"
	"net/netip"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Config contains all MTD services and cloud provider configs
type Config struct {
    MTD             mtdconf     `yaml:"mtd"`
    AWS             aws         `yaml:"aws"`
}

type mtdconf struct {
    Services        []service   `yaml:"services"`
}

type service struct {
    ID              customUUID  `yaml:"id"`
    ServiceID       string      `yaml:"cloud_id"`
    EntryIP         netip.Addr  `yaml:"entry_ip"`
    EntryPort       uint16      `yaml:"entry_port"`
    ServiceIP       netip.Addr  `yaml:"service_ip"`
    ServicePort     uint16      `yaml:"service_port"`
}

type customUUID uuid.UUID

type aws struct {
    Regions         []string    `yaml:"regions"`
    CredentialsPath string      `yaml:"credentials_path"`
}

func (u *customUUID) UnmarshalYAML(value *yaml.Node) error {
	id, err := uuid.Parse(value.Value)
	if err != nil {
		return err
	}
	*u = customUUID(id)
	return nil
}

// LoadConf loads config from a yaml file
func LoadConf(filename string) (Config, error) {
    var config Config

    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return config, err
    }

    err = yaml.Unmarshal([]byte(data), &config)
    if err != nil {
        return config, err
    }

    return config, nil
}