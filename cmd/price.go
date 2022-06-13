/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"gil/pkg/services/kube"
	"gil/pricer"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

type Giller struct {
	Provider pricer.ProviderInterface
	Cluster  kube.ClusterInterface
}

// priceCmd represents the price command
var priceCmd = &cobra.Command{
	Use:   "price",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var f pricer.ProviderInterface
		f = &pricer.ProviderAWS{}

		kc := kube.KubeClientConf{}
		c, err := kc.NewKubeClient()
		if err != nil {
			panic(err)
		}
		var k kube.ClusterInterface
		k = &kube.KubeConf{
			Client: c,
		}

		o := Giller{
			Provider: f,
			Cluster:  k,
		}

		pricedNodes := o.Provider.Nodes().Prices().List()
		clusterPricedNodes, err := o.Cluster.Prices(pricedNodes).List(
			cmd.Flag("namespace").Value.String(),
			cmd.Flag("label-selector").Value.String(),
		)
		if err != nil {
			panic(err)
		}

		for _, priced := range clusterPricedNodes {
			log.WithFields(log.Fields{
				"kind":              priced.Kind,
				"replicas":          priced.Replicas,
				"name":              priced.Name,
				"selector":          priced.Selector,
				"currency":          "USD",
				"requested_memory":  priced.RequestedMemory,
				"requested_cpu_mil": priced.RequestedCPUMil,
				"price_deployment":  priced.PricedDeployment,
				// "price_pods":       priced.PricedPod,
			}).Info("Estimated costs")
		}
	},
}

func init() {
	rootCmd.AddCommand(priceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// priceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	priceCmd.Flags().StringP("region", "r", "sa-east-1", "Price region where the instances are based")
}
