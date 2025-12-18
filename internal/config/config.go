package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database   DatabaseConfig
	Firebase   FirebaseConfig
	Cloudinary CloudinaryConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	MONGODB_URI string
	MONGODB_DB string
}

type FirebaseConfig struct {
	CredentialsPath string
}

type CloudinaryConfig struct {
	URL string
}

// Load loads configuration from .env file and environment variables
func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Warning: failed to read .env file, relying on system envs")
	}

	viper.SetDefault("PORT", "8080")

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			MONGODB_URI: getEnv("MONGODB_URI", ""),
			MONGODB_DB:  getEnv("MONGODB_DB", "rally_db"),
		},
		Firebase: FirebaseConfig{
			// In Cloud Run, this should be left empty ("")
			CredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		},
		Cloudinary: CloudinaryConfig{
			URL: getEnv("CLOUDINARY_URL", ""),
		},
	}

	return cfg
}

// getEnv is a helper for viper
func getEnv(key, defaultValue string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return defaultValue
}
