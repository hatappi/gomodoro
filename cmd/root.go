// Package cmd has rootCmd defined
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hatappi/gomodoro/config"
	"github.com/hatappi/gomodoro/logger"
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLog)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gomodoro/config.yaml)")

	rootCmd.PersistentFlags().String("log-file", "", "log file (default is $HOME/.gomodoro/gomodoro.log)")
	err := viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p, err := homedir.Expand("~/.gomodoro/gomodoro.sock")
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
		home, err := homedir.Expand("~/.gomodoro")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName("config.yaml")
	}

	viper.SetEnvPrefix("GOMODORO")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
}

func initLog() {
	config, err := config.GetConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p := config.LogFile
	if p == "" {
		p, err = homedir.Expand("~/.gomodoro/gomodoro.log")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err = os.MkdirAll(filepath.Dir(p), 0750); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logfile, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	logger.SetOutput(logfile)
	logger.SetLogLevel(logger.DebugLevel)
}
