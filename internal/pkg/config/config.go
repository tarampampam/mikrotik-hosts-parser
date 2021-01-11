// Package config contains application configuration struct with the most useful functions.
package config

import (
	"io/ioutil"

	"github.com/a8m/envsubst"
	"gopkg.in/yaml.v2"
)

// Config is main application configuration.
type Config struct {
	Sources []source `yaml:"sources"`

	Cache struct {
		File struct {
			DirPath string `yaml:"dir"`
		} `yaml:"files"`
		LifetimeSec uint32 `yaml:"lifetime_sec"`
	} `yaml:"cache"`

	RouterScript struct {
		Redirect struct {
			Address string `yaml:"address"`
		} `yaml:"redirect"`
		Exclude struct {
			Hosts []string `yaml:"hosts"`
		} `yaml:"exclude"`
		Comment            string `yaml:"comment"`
		MaxSourcesCount    uint16 `yaml:"max_sources"`
		MaxSourceSizeBytes uint32 `yaml:"max_source_size"`
	} `yaml:"router_script"`
}

type source struct {
	URI              string `yaml:"uri"`
	Name             string `yaml:"name"`
	Description      string `yaml:"description"`
	EnabledByDefault bool   `yaml:"enabled"`
	RecordsCount     uint   `yaml:"count"` // approximate quantity
}

// AddSource into sources list.
func (cfg *Config) AddSource(uri, name, description string, enabledByDefault bool, recordsCount uint) {
	cfg.Sources = append(cfg.Sources, source{
		URI:              uri,
		Name:             name,
		Description:      description,
		EnabledByDefault: enabledByDefault,
		RecordsCount:     recordsCount,
	})
}

// FromYaml configures itself using YAML content.
func (cfg *Config) FromYaml(in []byte, expandEnv bool) error {
	if expandEnv {
		parsed, err := envsubst.Bytes(in)
		if err != nil {
			return err
		}

		in = parsed
	}

	if err := yaml.UnmarshalStrict(in, cfg); err != nil {
		return err
	}

	return nil
}

// FromYaml creates new config instance using YAML-structured content.
func FromYaml(in []byte, expandEnv bool) (*Config, error) {
	config := &Config{}

	if err := config.FromYaml(in, expandEnv); err != nil {
		return nil, err
	}

	return config, nil
}

// FromYamlFile creates new config instance using YAML file.
func FromYamlFile(filename string, expandEnv bool) (*Config, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return FromYaml(bytes, expandEnv)
}
