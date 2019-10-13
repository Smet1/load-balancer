package health

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	stringDuration := ""
	err := unmarshal(&stringDuration)
	if err != nil {
		return err
	}

	d.Duration, err = time.ParseDuration(stringDuration)
	return err
}

type PathAndResult struct {
	Addr           string `yaml:"addr"`
	IP             string `yaml:"ip"`
	ExpectedStatus int    `yaml:"expected_status"`
}

type Config struct {
	Services        []PathAndResult `yaml:"services"`
	NginxConfigPath string          `yaml:"nginx_config_path"`
	Ticker          Duration        `yaml:"ticker"`
	PatternAddr     string          `yaml:"pattern_addr"`
	Before          string          `yaml:"before"`
	After           string          `yaml:"after"`
}

func ReadConfig(fileName string, config interface{}) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return errors.Wrap(err, "cant read config file")
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return errors.Wrap(err, "cant parse config")
	}

	return nil
}
