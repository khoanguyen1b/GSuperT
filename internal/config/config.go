package config

import (
	"log"
	"github.com/spf13/viper"
)

type Config struct {
	AppPort       string `mapstructure:"APP_PORT"`
	DBHost        string `mapstructure:"DB_HOST"`
	DBPort        string `mapstructure:"DB_PORT"`
	DBUser        string `mapstructure:"DB_USER"`
	DBPassword    string `mapstructure:"DB_PASSWORD"`
	DBName        string `mapstructure:"DB_NAME"`
	DBSSLMode     string `mapstructure:"DB_SSLMODE"`
	JWTSecret     string `mapstructure:"JWT_SECRET"`
	JWTRefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
	AccessTokenExp  int    `mapstructure:"ACCESS_TOKEN_EXP_MINUTES"`
	RefreshTokenExp int    `mapstructure:"REFRESH_TOKEN_EXP_DAYS"`
	SMTPHost        string `mapstructure:"SMTP_HOST"`
	SMTPPort        string `mapstructure:"SMTP_PORT"`
	SMTPUser        string `mapstructure:"SMTP_USER"`
	SMTPPass        string `mapstructure:"SMTP_PASS"`
	SMTPFrom        string `mapstructure:"SMTP_FROM"`
	SMTPFromName    string `mapstructure:"SMTP_FROM_NAME"`
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		log.Fatal("Unable to decode into struct: ", err)
	}

	return config
}
