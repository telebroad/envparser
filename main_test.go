package flag_env_to_struct

import (
	"os"
	"testing"
)

type db struct {
	URI                string `flag:"uri,usage:db connection uri" env:"URI"`
	MaxIdleConnections int    `flag:"max-idle-connections" env:"MAX_IDLE_CONNECTIONS"`
	MaxOpenConnections int    `flag:"max-open-connections" env:"MAX_OPEN_CONNECTIONS"`
	ConnMaxIdleTime    int    `flag:"conn-max-idle-time" env:"CONN_MAX_IDLE_TIME"`
	ConnMaxLifeTime    int    `flag:"conn-max-life-time" env:"CONN_MAX_LIFE_TIME"`
}

type State struct {
	LogDB      db     `flag:"log-db" env:"LOG_DB"`
	CustomerId int64  `flag:"customer-id, usage:on what customer id it should work on" env:"CUSTOMER_ID"`
	Start      int    `flag:"start-line, usage:some id to start" env:"START"`
	End        int    `flag:"end-line, usage:some id to end" env:"END"`
	EnvFile    string `flag:"env-file, default:.env,usage:to change the default env file to load" env:"-"`
}

func TestSetUpFlags(t *testing.T) {
	state := State{}
	err := SetUpFlags(".", &state)
	if err != nil {
		t.Errorf("error: %s", err)
	}

	t.Logf("%+v", state)
}

func TestSetUpEnv(t *testing.T) {
	os.Setenv("CUSTOMER_ID", "1")
	os.Setenv("START", "2")
	os.Setenv("END", "3")
	os.Setenv("LOG_DB.URI", "uri")
	os.Setenv("LOG_DB.MAX_IDLE_CONNECTIONS", "1")
	os.Setenv("LOG_DB.MAX_OPEN_CONNECTIONS", "2")
	os.Setenv("LOG_DB.CONN_MAX_IDLE_TIME", "3")
	os.Setenv("LOG_DB.CONN_MAX_LIFE_TIME", "4")

	state := State{}
	err := SetUpEnv(".", &state)
	if err != nil {
		t.Errorf("error: %s", err)
	}

	t.Logf("%+v", state)

}
