package state

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)


type Config struct {
    AWS struct {
        Regions     []string `yaml:"regions"`
    } `yaml:"aws"`
}

func Load_conf(filename string) (Config, error) {
    var config Config

    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return config, err
    }

    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return config, err
    }

    return config, nil
}