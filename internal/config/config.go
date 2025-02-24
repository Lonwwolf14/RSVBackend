package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL string
}

func Read() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	DBConnectionString := os.Getenv("DBURL")
	if DBConnectionString == "" {
		return nil, nil
	}

	var configData Config
	configData.DBURL = DBConnectionString
	return &configData, nil

}
