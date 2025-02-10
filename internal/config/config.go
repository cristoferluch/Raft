package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port                   int      `yaml:"port"`
	HeartbeatInterval      string   `yaml:"heartbeat_interval"`
	LeaderHeartbeatTimeout string   `yaml:"leader_heartbeat_timeout"`
	Node                   string   `yaml:"node"`
	Peers                  []string `yaml:"peers"`
}

func LoadConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
