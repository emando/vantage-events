// Copyright © 2020 Emando B.V.

package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	Use:   "eventrecorder",
	Short: "Event Recorder for Vantage event streaming.",
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.eventrecorder.yaml)")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debugging")
	rootCmd.PersistentFlags().String("file", "log.json", "file")
	rootCmd.PersistentFlags().String("host", "events.emandovantage.com", "Vantage Events Server host")

	viper.BindPFlags(rootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".eventrecorder")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("vantage")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("using config file:", viper.ConfigFileUsed())
	}
}
