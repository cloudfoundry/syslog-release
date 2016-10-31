package blackbox

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/concourse/blackbox/syslog"
	"gopkg.in/yaml.v2"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var num int64
	if err := unmarshal(&num); err == nil {
		*d = Duration(num)
		return nil
	}

	var str string
	if err := unmarshal(&str); err != nil {
		return errors.New("invalid duration; must be string or number")
	}

	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	*d = Duration(duration)

	return nil
}

type SyslogSource struct {
	Path string `yaml:"path"`
	Tag  string `yaml:"tag"`
}

type SyslogConfig struct {
	Destination syslog.Drain `yaml:"destination"`
	SourceDir   string       `yaml:"source_dir"`
}

type Config struct {
	Hostname string `yaml:"hostname"`

	Syslog SyslogConfig `yaml:"syslog"`
}

func LoadConfig(path string) (*Config, error) {
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config

	if err := yaml.Unmarshal(configFile, &config); err != nil {
		return nil, err
	}

	if config.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		config.Hostname = hostname
	}

	return &config, nil
}
