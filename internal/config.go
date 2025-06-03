package internal

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	APIKey               string `yaml:"api_key"`
	BaseURL              string `yaml:"health_base_url"`
	DefaultCheckInterval int    `yaml:"ping_interval"`
	DefaultGrace         int    `yaml:"default_grace"`
}

type Task struct {
	Name     string   `yaml:"name"`
	Slug     string   `yaml:"slug"`           // jika belum ada, tambahkan
	UUID     string   `yaml:"uuid,omitempty"` // boleh kosong kalau baru akan dibuat
	Shell    string   `yaml:"shell"`
	Interval int      `yaml:"interval,omitempty"`
	Grace    int      `yaml:"grace,omitempty"`
	Tags     []string `yaml:"tags"`
}

type Config struct {
	Global GlobalConfig `yaml:"global"`
	Tasks  []Task       `yaml:"tasks"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}
