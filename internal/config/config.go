package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port         string
	DatabaseURL  string
	FirebaseCred string
	JWTSecret    string
	Env          string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Warning: failed to read .env file, relying on system envs")
	}

	viper.SetDefault("PORT", "8080")

	return &Config{
		Port:         viper.GetString("PORT"),
		DatabaseURL:  viper.GetString("DATABASE_URL"),
		FirebaseCred: viper.GetString("FIREBASE_CRED"),
		JWTSecret:    viper.GetString("JWT_SECRET"),
		Env:          viper.GetString("ENV"),
	}
}
