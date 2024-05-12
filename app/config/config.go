package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	AppPort           string
	LogEnv            string
	CtxTimeOut        time.Duration
	ApiPath           ApiPath
	RetryForCheck     RetryForCheck
	HTTPClientTimeOut time.Duration
}

type ApiPath struct {
	BroadCastTxnPath string
	CheckTxnPath     string
	HealthCheckPath  string
	BroadCastExtPath string
	PendingCheck     string
}

type RetryForCheck struct {
	RetryTimes int
	RetryDelay time.Duration
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
	config.CtxTimeOut = viper.GetDuration("context_timeout_graceful")
	config.HTTPClientTimeOut = viper.GetDuration("http_client_timeout")
	config.RetryForCheck.RetryTimes = viper.GetInt("retry_for_check.retry_times")
	config.RetryForCheck.RetryDelay = viper.GetDuration("retry_for_check.retry_repeat_delay")
	config.ApiPath.BroadCastTxnPath = viper.GetString("api_path.post-txn")
	config.ApiPath.CheckTxnPath = viper.GetString("api_path.get-txn")
	config.ApiPath.HealthCheckPath = viper.GetString("api_path.health-check")
	config.ApiPath.BroadCastExtPath = viper.GetString("api_path.broadcast-ext-txn-path")
	config.ApiPath.PendingCheck = viper.GetString("api_path.check-ext-txn-pending")
	log.Println(config)

	return config
}
