package common

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Config struct {
	DataBase DataBaseConfig `json:"data_base"`
	Redis    RedisConfig    `json:"redis"`
	Server   ServerConfig   `json:"server"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DataBaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func GetConfig(filePath string) (*Config, error) {
	rawConfig, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.New("can not get config: " + err.Error())
	}

	config := Config{}
	err = json.Unmarshal(rawConfig, &config)
	if err != nil {
		return nil, errors.New("can not get config: " + err.Error())
	}
	return &config, nil
}
