package main

import (
	"github.com/telebroad/envparser"
	"log/slog"
	"os"
)

type db struct {
	URI                string `json:"uri" flag:"uri,usage:db connection uri" env:"URI"`
	MaxIdleConnections int    `json:"maxIdleConnections" flag:"max-idle-connections" env:"MAX_IDLE_CONNECTIONS"`
	MaxOpenConnections int    `json:"maxOpenConnections" flag:"max-open-connections" env:"MAX_OPEN_CONNECTIONS"`
	ConnMaxIdleTime    int    `json:"connMaxIdleTime" flag:"conn-max-idle-time" env:"CONN_MAX_IDLE_TIME"`
	ConnMaxLifeTime    int    `json:"connMaxLifeTime" flag:"conn-max-life-time" env:"CONN_MAX_LIFE_TIME"`
}

type State struct {
	LogDB      db     `json:"logDB" flag:"log-db" env:"LOG_DB"`
	CustomerId int64  `json:"customerId" flag:"customer-id, usage:on what customer id it should work on" env:"CUSTOMER_ID"`
	Start      int    `json:"start" flag:"start-line, usage:some id to start" env:"START"`
	End        int    `json:"end" flag:"end-line, usage:some id to end" env:"END"`
	EnvFile    string `json:"envFile" flag:"env-file, default:.env,usage:to change the default env file to load" env:"-"`
}

func setUpEnvTest(log *slog.Logger) {

	log = log.With("test", "SetUpFlagEnv")
	os.Setenv("CUSTOMER_ID", "1")
	os.Setenv("START", "2")
	os.Setenv("END", "3")
	os.Setenv("LOG_DB.URI", "user:password@tcp(localhost:3306)/dbname?parseTime=true")
	os.Setenv("LOG_DB.MAX_IDLE_CONNECTIONS", "1")
	os.Setenv("LOG_DB.MAX_OPEN_CONNECTIONS", "2")
	os.Setenv("LOG_DB.CONN_MAX_IDLE_TIME", "3")
	os.Setenv("LOG_DB.CONN_MAX_LIFE_TIME", "4")

	state := State{}
	err := envparser.SetUpFlagEnv(".", &state)
	if err != nil {
		log.Error("SetUpEnv returned an error", "error", err)
	}
	log.Debug("state", "state", state)
}

func main() {
	logOpt := &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}
	log := slog.New(slog.NewTextHandler(os.Stdout, logOpt))
	slog.SetDefault(log)
	envparser.SetLogger(log)
	defer log.Info("exiting")
	setUpEnvTest(log)
}
