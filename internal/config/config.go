// Package config gomodoro configuration
package config

import (
	"reflect"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/go-playground/validator/v10"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

const (
	// DefaultWorkSec default working second.
	DefaultWorkSec = 1500
	// DefaultShortBreakSec default short break second.
	DefaultShortBreakSec = 300
	// DefaultLongBreakSec default long break second.
	DefaultLongBreakSec = 900

	// DefaultLogFile default log file path.
	DefaultLogFile = "~/.gomodoro/gomodoro.log"

	// DefaultStorageDir is default storage directory.
	DefaultStorageDir = "~/.gomodoro"

	// DefaultAPITimeout default timeout for API operations in seconds.
	DefaultAPITimeout = 10
)

// Config config for gomodoro.
type Config struct {
	Pomodoro PomodoroConfig `mapstructure:"pomodoro"`
	Toggl    TogglConfig    `mapstructure:"toggl"`
	Color    ColorConfig    `mapstructure:"color"`
	Pixela   PixelaConfig   `mapstructure:"pixela"`
	LogFile  string         `mapstructure:"log_file"`
	LogLevel zapcore.Level  `mapstructure:"log_level"`
	API      APIConfig      `mapstructure:"api"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

// StorageConfig contains configuration options for storage.
type StorageConfig struct {
	Dir string `mapstructure:"dir"`
}

// APIConfig contains configuration options for the API server.
type APIConfig struct {
	Addr         string        `mapstructure:"addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// PomodoroConfig config for pomodoro.
type PomodoroConfig struct {
	WorkSec        int `mapstructure:"work_sec"        validate:"gt=0,lte=3600"`
	ShortBreakSec  int `mapstructure:"short_break_sec" validate:"gt=0,lte=3600"`
	LongBreakSec   int `mapstructure:"long_break_sec"  validate:"gt=0,lte=3600"`
	BreakFrequency int `mapstructure:"break_frequency" validate:"gte=2,lte=9"`
}

// TogglConfig config for Toggl.
type TogglConfig struct {
	Enable      bool   `mapstructure:"enable"`
	APIToken    string `mapstructure:"api_token"`
	ProjectID   int    `mapstructure:"project_id"`
	WorkspaceID int    `mapstructure:"workspace_id"`
}

// PixelaConfig is configuration for Pixela
// https://pixe.la/
type PixelaConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Token    string `mapstructure:"token"`
	UserName string `mapstructure:"user_name"`
	GraphID  string `mapstructure:"graph_id"`
}

// ColorConfig represents colors used within gomodoro.
type ColorConfig struct {
	Font                tcell.Color `mapstructure:"font"`
	Background          tcell.Color `mapstructure:"background"`
	SelectedLine        tcell.Color `mapstructure:"selected_line"`
	StatusBarBackground tcell.Color `mapstructure:"status_bar_background"`
	TimerPauseFont      tcell.Color `mapstructure:"timer_pause_font"`
	TimerWorkFont       tcell.Color `mapstructure:"timer_work_font"`
	TimerBreakFont      tcell.Color `mapstructure:"timer_break_font"`
	Cursor              tcell.Color `mapstructure:"cursor"`
}

// DefaultConfig get default config.
func DefaultConfig() *Config {
	return &Config{
		Toggl: TogglConfig{
			Enable: false,
		},
		Pomodoro: PomodoroConfig{
			WorkSec:        DefaultWorkSec,
			ShortBreakSec:  DefaultShortBreakSec,
			LongBreakSec:   DefaultLongBreakSec,
			BreakFrequency: 2, //nolint:mnd
		},
		LogFile: DefaultLogFile,
		Color: ColorConfig{
			Font:                tcell.ColorDarkSlateGray,
			Background:          tcell.ColorWhite,
			SelectedLine:        tcell.ColorBlue,
			StatusBarBackground: tcell.ColorBlack,
			TimerPauseFont:      tcell.ColorDarkOrange,
			TimerWorkFont:       tcell.ColorGreen,
			TimerBreakFont:      tcell.ColorBlue,
			Cursor:              tcell.ColorGreen,
		},
		API: APIConfig{
			Addr:         "localhost:8080",
			ReadTimeout:  time.Second * DefaultAPITimeout,
			WriteTimeout: time.Second * DefaultAPITimeout,
		},
		Storage: StorageConfig{
			Dir: DefaultStorageDir,
		},
	}
}

// GetConfig get Config.
func GetConfig() (*Config, error) {
	c := DefaultConfig()

	err := viper.Unmarshal(&c,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				tcellColorDecodeHook(),
				zapcoreLevelDecodeHook(),
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

	// Expand each file

	if c.LogFile, err = homedir.Expand(c.LogFile); err != nil {
		return nil, err
	}

	if c.Storage.Dir, err = homedir.Expand(c.Storage.Dir); err != nil {
		return nil, err
	}

	return c, nil
}

func tcellColorDecodeHook() mapstructure.DecodeHookFunc {
	return func(_ reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(tcell.Color(0)) {
			return data, nil
		}

		if str, ok := data.(string); ok {
			return tcell.GetColor(str), nil
		}

		return data, nil
	}
}

func zapcoreLevelDecodeHook() mapstructure.DecodeHookFunc {
	return func(_ reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(zapcore.Level(0)) {
			return data, nil
		}

		if str, ok := data.(string); ok {
			lvl := new(zapcore.Level)
			if err := lvl.UnmarshalText([]byte(str)); err != nil {
				return nil, err
			}

			return lvl, nil
		}

		return data, nil
	}
}
