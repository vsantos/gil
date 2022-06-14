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
var (
	priceCmd = &cobra.Command{
		Use:   "price",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			var f pricer.ProviderInterface
			f = &pricer.ProviderAWS{}

			kc := kube.KubeClientConf{}
			c, err := kc.NewKubeClient()
			if err != nil {
				log.Fatal(err)
			}

			var k kube.ClusterInterface
			k = &kube.KubeConf{
				Client: c,
				Region: cmd.Flag("region").Value.String(),
			}

			o := Giller{
				Provider: f,
				Cluster:  k,
			}

			pricedNodes := o.Provider.Nodes().Prices().List()

			i, err := o.Cluster.Prices(pricedNodes)
			if err != nil {
				log.Fatal(err)
			}

			clusterPricedNodes, err := i.List(
				cmd.Flag("namespace").Value.String(),
				cmd.Flag("label-selector").Value.String(),
			)
			if err != nil {
				log.Fatal(err)
			}

			if len(clusterPricedNodes) == 0 {
				log.Fatal("could not find any deployments returned by filter `-l '%s'`")
			}

			if len(clusterPricedNodes) > 0 {
				for _, priced := range clusterPricedNodes {
					// show associated pods if needed
					showPods, _ := cmd.Flags().GetBool("show-pods")
					if !showPods {
						priced.Deployment.Pods = []kube.ClusterPodPrice{}
					}

					logFields := log.Fields{
						"selector":          priced.Selector,
						"currency":          "USD",
						"requested_memory":  priced.RequestedMemory,
						"requested_cpu_mil": priced.RequestedCPUMil,
						"deployment":        priced.Deployment,
					}
					log.WithFields(logFields).Info("Estimated costs")
				}
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(priceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// priceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	priceCmd.Flags().StringP("region", "r", "sa-east-1", "Price region where the instances are based")
	priceCmd.Flags().Bool("show-pods", false, "List individual pods price from a Deployment")
}
