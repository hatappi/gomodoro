package config

import (
	"github.com/BurntSushi/toml"
	"github.com/hatappi/gomodoro/libs/toggl/config"
)

type Config struct {
	Toggl         *toggl.Config
	AppDir        string
	LongBreakSec  int
	ShortBreakSec int
	WorkSec       int
}

func LoadConfig(path string) *Config {
	var conf Config

	if _, err := toml.DecodeFile(path, &conf); err != nil {
		panic(err)
	}
	return &conf
}
