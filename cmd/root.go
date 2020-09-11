// Package cmd has rootCmd defined
package cmd

import (
	"context"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"

	"github.com/hatappi/go-kit/log"

	"github.com/hatappi/gomodoro/internal/config"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "gomodoro",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx := context.Background()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gomodoro/config.yaml)")

	rootCmd.PersistentFlags().String("log-file", "", "log file (default is $HOME/.gomodoro/gomodoro.log)")
	err := viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().String("log-level", "error", "log Level (default is error)")
	err = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p, err := homedir.Expand(config.DefaultUnixDomainScoketPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rootCmd.PersistentFlags().String("unix-domain-socket-path", p, "unix domain socket path")
	err = viper.BindPFlag("unix_domain_socket_path", rootCmd.PersistentFlags().Lookup("unix-domain-socket-path"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configPath, err := homedir.Expand("~/.gomodoro/config.yaml")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.SetConfigFile(configPath)
	}

	viper.SetEnvPrefix("GOMODORO")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
}

func initLogger() {
	conf, err := config.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config. %s\n", err)
		os.Exit(1)
	}

	p := conf.LogFile
	if p == "" {
		p = config.DefaultLogFile
	}

	p, err = homedir.Expand(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get log file path. %s\n", err)
		os.Exit(1)
	}

	level := zapcore.Level(0)
	if err = level.UnmarshalText([]byte(conf.LogLevel)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal level. %s\n", err)
		os.Exit(1)
	}

	logger, err := log.New(
		"gomodoro",
		log.WithOutputPaths([]string{p}),
		log.WithErrorOutputPaths([]string{p}),
		log.WithLevel(level),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init Logger. %s\n", err)
		os.Exit(1)
	}

	log.SetDefaultLogger(logger)
}
