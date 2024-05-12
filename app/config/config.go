package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	AppPort    string
	LogEnv     string
	CtxTimeOut time.Duration
}

func ConfigLoader() *AppConfig {
	config := new(AppConfig)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error loading config file: %s", err.Error()))
	}

	config.AppPort = viper.GetString("app.port")
	config.LogEnv = viper.GetString("logger.env")
	config.CtxTimeOut = viper.GetDuration("context_timeout")
	log.Println(config)

	return config
}
