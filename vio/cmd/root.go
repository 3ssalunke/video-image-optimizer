package cmd

import (
	"log"

	"github.com/3ssalunke/vio/vio/config"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/spf13/cobra"
)

var (
	configFile string
	rootCmd    = &cobra.Command{
		Use: "vio",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
	rootCmd.AddCommand(serveCmd)
}

func initConfig() {
	var err error

	if configFile != "" {
		err = cleanenv.ReadConfig(configFile, &config.App)
	} else {
		err = cleanenv.ReadConfig("etc/vio/config.yml", &config.App)
	}

	if err != nil {
		log.Fatalf("config error: %s", err)
	}
}
