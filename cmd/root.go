/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gil",
	Short: "Get estimated node costs based on CPU/Mem requests",
	Long: `Gil is a binary that allows to retrieve your deployments estimated costs
based on it's CPU and Memory requests using label-selector as filter. With this
information gil is able to calculate the percent of usage from it's pods within a
node. The method of calculation is very simple and not covers limits or fluctuations.

You can also retrieve individual pod costs along with it's deployment by '--show-labels'

The only supported provider for now is 'AWS'. Which means the Gil is only able to estimate
costs from pods running within ec2 instances.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gil.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringP("provider", "p", "aws", "Cloud provider to get node price list from")
	rootCmd.PersistentFlags().StringP("label-selector", "l", "", "Selector (label query) to filter on, supports '='.(e.g. -l key1=value1)")
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "Namespace to filter resources price from")

	rootCmd.MarkPersistentFlagRequired("label-selector")
	rootCmd.MarkPersistentFlagRequired("namespace")
}
