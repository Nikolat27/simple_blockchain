package utils

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		file, err := os.Create(".env")
		if err != nil {
			return fmt.Errorf("error creating .env file: %v", err)
		}

		defer file.Close()
	}

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load the .env variables")
	}

	return nil
}
