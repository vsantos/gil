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
	i, err := o.Cluster.Prices(pricedNodes)
	if err != nil {
		log.Fatal(err)
	}

	clusterPricedNodes, err := i.List("namespace", "key1=value1")
	if err != nil {
		log.Panic(err)
	}
	if len(clusterPricedNodes) > 0 {
		for _, priced := range clusterPricedNodes {
			// show associated pods if needed
			showPods := true
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
}
