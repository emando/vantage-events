// Copyright © 2019 Emando B.V.

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile string
	debug   bool
	logger  *zap.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aggregator",
	Short: "Event Aggregator for Vantage event streaming.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		if viper.GetBool("debug") {
			logger, err = zap.NewDevelopment()
		} else {
			logger, err = zap.NewProduction()
		}
		if err != nil {
			log.Fatalf("failed to initialize zap (%v)", err)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		logger.Sync()
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aggregator.yaml)")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debugging")
	rootCmd.PersistentFlags().String("driver", "nats", "driver (nats)")

	rootCmd.PersistentFlags().String("nats-url", "nats://events.emandovantage.com:4222", "NATS Streaming Server URL")
	rootCmd.PersistentFlags().String("nats-username", "", "NATS username")
	rootCmd.PersistentFlags().String("nats-password", "", "NATS password")
	rootCmd.PersistentFlags().Bool("nats-tls", true, "use TLS for NATS")
	rootCmd.PersistentFlags().String("nats-cluster-id", "vantage", "NATS cluster ID")
	rootCmd.PersistentFlags().String("nats-client-id", "aggregator", "NATS client ID")

	viper.BindPFlags(rootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".aggregator")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("vantage")
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("using config file:", viper.ConfigFileUsed())
	}
}
