// Package cmd has rootCmd defined
package cmd

import (
	"context"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/go-kit/log/zap"

	"github.com/hatappi/gomodoro/internal/config"
)

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "~/.gomodoro/config.yaml", "config file")

	rootCmd.PersistentFlags().String("log-file", config.DefaultLogFile, "log file")
	cobra.CheckErr(viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file")))

	rootCmd.PersistentFlags().String("log-level", "error", "log Level (default is error)")
	cobra.CheckErr(viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level")))

	rootCmd.PersistentFlags().String("unix-domain-socket-path", config.DefaultUnixDomainScoketPath, "unix domain socket path")
	cobra.CheckErr(viper.BindPFlag("unix_domain_socket_path", rootCmd.PersistentFlags().Lookup("unix-domain-socket-path")))
}

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

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configPath, err := homedir.Expand(cfgFile)
	cobra.CheckErr(err)
	viper.SetConfigFile(configPath)

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

	logger, err := zap.NewLogger("gomodoro",
		zap.WithOutputPaths([]string{conf.LogFile}),
		zap.WithErrorOutputPaths([]string{conf.LogFile}),
		zap.WithLevel(conf.LogLevel),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init Logger. %s\n", err)
		os.Exit(1)
	}

	log.SetDefaultLogger(logger)
}
