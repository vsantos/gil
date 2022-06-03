package main

import (
	"fmt"
	"gil/pricer"
	"gil/services"
)

type Giller struct {
	Provider pricer.ProviderInterface
	Cluster  services.ClusterInterface
}

func main() {
	var f pricer.ProviderInterface
	f = &pricer.ProviderAWS{}

	kc := services.KubeClientConf{}
	c, err := kc.NewKubeClient()
	if err != nil {
		panic(err)
	}
	var k services.ClusterInterface
	k = &services.KubeConf{
		Client: c,
	}

	o := Giller{
		Provider: f,
		Cluster:  k,
	}

	pricedNodes := o.Provider.Nodes().Prices().List()
	clusterPricedNodes, err := o.Cluster.Pods(pricedNodes).Prices().List("kong-system", "app=ingress-kong")
	if err != nil {
		panic(err)
	}

	fmt.Println(clusterPricedNodes)
}
