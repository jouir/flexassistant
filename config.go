package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config to receive settings from the configuration file
type Config struct {
	DatabaseFile   string         `yaml:"database-file"`
	MaxBlocks      int            `yaml:"max-blocks"`
	MaxPayments    int            `yaml:"max-payments"`
	Pools          []PoolConfig   `yaml:"pools"`
	Miners         []MinerConfig  `yaml:"miners"`
	TelegramConfig TelegramConfig `yaml:"telegram"`
}

// PoolConfig to store Pool configuration
type PoolConfig struct {
	Coin         string `yaml:"coin"`
	EnableBlocks bool   `yaml:"enable-blocks"`
}

// MinerConfig to store Miner configuration
type MinerConfig struct {
	Address              string `yaml:"address"`
	EnableBalance        bool   `yaml:"enable-balance"`
	EnablePayments       bool   `yaml:"enable-payments"`
	EnableOfflineWorkers bool   `yaml:"enable-offline-workers"`
}

// TelegramConfig to store Telegram configuration
type TelegramConfig struct {
	Token       string `yaml:"token"`
	ChatID      int64  `yaml:"chat-id"`
	ChannelName string `yaml:"channel-name"`
}

// NewConfig creates a Config with default values
func NewConfig() *Config {
	return &Config{
		DatabaseFile: AppName + ".db",
	}
}

// ReadFile reads and parses a YAML configuration file to override default values
func (c *Config) ReadFile(filename string) (err error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return err
	}

	return nil
}
