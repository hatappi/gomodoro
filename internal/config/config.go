// Package config gomodoro configuration
package config

import (
	"reflect"

	"github.com/gdamore/tcell"
	"github.com/go-playground/validator/v10"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
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
	Color                ColorConfig    `mapstructure:"color"`
	LogFile              string         `mapstructure:"log_file"`
	LogLevel             string         `mapstructure:"log_level"`
	TaskFile             string         `mapstructure:"task_file"`
	UnixDomainScoketPath string         `mapstructure:"unix_domain_socket_path"`
}

// ExpandTaskFile get expand task file
func (c *Config) ExpandTaskFile() (string, error) {
	p, err := homedir.Expand(c.TaskFile)
	if err != nil {
		return "", err
	}
	return p, nil
}

// ExpandLogFile get expand log file
func (c *Config) ExpandLogFile() (string, error) {
	p, err := homedir.Expand(c.LogFile)
	if err != nil {
		return "", err
	}
	return p, nil
}

// ExpandUnixDomainSocketPath get expand UnixDomainScoketPath
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

type ColorConfig struct {
	Font                tcell.Color `mapstructure:"font"`
	Background          tcell.Color `mapstructure:"background"`
	SelectedLine        tcell.Color `mapstructure:"selected_line"`
	StatusBarBackground tcell.Color `mapstructure:"status_bar_background"`
}

func defaultConfig() *Config {
	return &Config{
		Color: ColorConfig{
			Font:                tcell.ColorDarkSlateGray,
			Background:          tcell.ColorWhite,
			SelectedLine:        tcell.ColorBlue,
			StatusBarBackground: tcell.ColorBlack,
		},
	}
}

// GetConfig get Config
func GetConfig() (*Config, error) {
	c := defaultConfig()

	err := viper.Unmarshal(&c,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				TcellColorDecodeHook(),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func TcellColorDecodeHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(tcell.Color(0)) {
			return data, nil
		}

		if str, ok := data.(string); ok {
			return tcell.GetColor(str), nil
		}

		return data, nil
	}
}
