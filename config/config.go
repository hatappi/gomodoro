// Package config gomodoro configuration
package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config config for gomodoro
type Config struct {
	Pomodoro PomodoroConfig `mapstructure:"pomodoro"`
	LogFile  string         `mapstructure:"log_file"`
	TaskFile string         `mapstructure:"task_file"`
}

// PomodoroConfig config for pomodoro
type PomodoroConfig struct {
	WorkSec       int `mapstructure:"work_sec" validate:"gt=0,lte=3600"`
	ShortBreakSec int `mapstructure:"short_break_sec" validate:"gt=0,lte=3600"`
	LongBreakSec  int `mapstructure:"long_break_sec" validate:"gt=0,lte=3600"`
}

// GetConfig get Config
func GetConfig() (*Config, error) {
	var c Config
	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
