package main

import (
	"gil/cmd"
	kube "gil/pkg/services/kube"
	"gil/pricer"
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {
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

	o := cmd.Giller{
		Provider: f,
		Cluster:  k,
	}

	pricedNodes := o.Provider.Nodes().Prices().List()
	clusterPricedNodes, err := o.Cluster.Prices(pricedNodes).List("ext", "squad=psm-transactions")
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
}
