// Package config gomodoro configuration
package config

import (
	"github.com/go-playground/validator/v10"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	// DefaultWorkSec default working second
	DefaultWorkSec = 1500
	// DefaultShortBreakSec default short break second
	DefaultShortBreakSec = 300
	// DefaultLongBreakSec default long break second
	DefaultLongBreakSec = 900

	// DefaultLogFile default log file path
	DefaultLogFile = "~/.gomodoro/gomodoro.log"
	// DefaultTaskFile default task file path
	DefaultTaskFile = "~/.gomodoro/tasks.yaml"
	// DefaultUnixDomainScoketPath default unix domain socket file path
	DefaultUnixDomainScoketPath = "~/.gomodoro/gomodoro.sock"
)

// Config config for gomodoro
type Config struct {
	Pomodoro             PomodoroConfig `mapstructure:"pomodoro"`
	Toggl                TogglConfig    `mapstructure:"toggl"`
	LogFile              string         `mapstructure:"log_file"`
	TaskFile             string         `mapstructure:"task_file"`
	UnixDomainScoketPath string         `mapstructure:"unix_domain_socket_path"`
}

func (c *Config) ExpandTaskFile() (string, error) {
	p, err := homedir.Expand(c.TaskFile)
	if err != nil {
		return "", err
	}
	return p, nil
}

func (c *Config) ExpandLogFile() (string, error) {
	p, err := homedir.Expand(c.LogFile)
	if err != nil {
		return "", err
	}
	return p, nil
}

func (c *Config) ExpandUnixDomainSocketPath() (string, error) {
	p, err := homedir.Expand(c.UnixDomainScoketPath)
	if err != nil {
		return "", err
	}
	return p, nil
}

// PomodoroConfig config for pomodoro
type PomodoroConfig struct {
	WorkSec       int `mapstructure:"work_sec" validate:"gt=0,lte=3600"`
	ShortBreakSec int `mapstructure:"short_break_sec" validate:"gt=0,lte=3600"`
	LongBreakSec  int `mapstructure:"long_break_sec" validate:"gt=0,lte=3600"`
}

// TogglConfig config for Toggl
type TogglConfig struct {
	APIToken  string `mapstructure:"api_token"`
	ProjectID int    `mapstructure:"project_id"`
}

// Enable confirm toggl client is enable
func (tc TogglConfig) Enable() bool {
	return tc.APIToken != "" && tc.ProjectID != 0
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
