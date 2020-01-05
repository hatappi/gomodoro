// Package config gomodoro configuration
package config

import (
	"github.com/spf13/viper"
)

// Config config for gomodoro
type Config struct {
	Pomodoro PomodoroConfig `mapstructure:"pomodoro"`
}

// PomodoroConfig config for pomodoro
type PomodoroConfig struct {
	WorkSec       int `mapstructure:"work_sec"`
	ShortBreakSec int `mapstructure:"short_break_sec"`
	LongBreakSec  int `mapstructure:"long_break_sec"`
}

// GetConfig get Config
func GetConfig() (*Config, error) {
	var c Config
	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
