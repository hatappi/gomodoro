// Package cmd has rootCmd defined
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/go-kit/log/zap"

	"github.com/hatappi/gomodoro/internal/config"
)

var cfgFile string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx := context.Background()

	cobra.OnInitialize(initConfig, initLogger)

	rootCmd := newRootCmd()
	rootCmd.AddCommand(
		newVersionCmd(),
		newStartCmd(),
		newRemainCmd(),
		newInitCmd(),
		newAddTaskCmd(),
		newServeCmd(),
	)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "gomodoro",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "~/.gomodoro/config.yaml", "config file")

	rootCmd.PersistentFlags().String("log-file", config.DefaultLogFile, "log file")
	cobra.CheckErr(viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file")))

	rootCmd.PersistentFlags().String("log-level", "error", "log Level (default is error)")
	cobra.CheckErr(viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level")))

	rootCmd.PersistentFlags().String(
		"unix-domain-socket-path",
		config.DefaultUnixDomainScoketPath,
		"unix domain socket path",
	)
	cobra.CheckErr(viper.BindPFlag("unix_domain_socket_path", rootCmd.PersistentFlags().Lookup("unix-domain-socket-path")))

	return rootCmd
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

	logDir := filepath.Dir(conf.LogFile)

	_, err = os.Stat(logDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(conf.LogFile), 0o750) //nolint:mnd
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create log directory. %s\n", err)
			os.Exit(1)
		}
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
