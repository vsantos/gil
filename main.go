package main

import (
	"fmt"
	kube "gil/pkg/services/kube"
	"gil/pricer"
)

type Giller struct {
	Provider pricer.ProviderInterface
	Cluster  kube.ClusterInterface
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

	o := Giller{
		Provider: f,
		Cluster:  k,
	}

	pricedNodes := o.Provider.Nodes().Prices().List()
	clusterPricedNodes, err := o.Cluster.Prices(pricedNodes).List("ext", "app=worker-sweep-account")
	if err != nil {
		panic(err)
	}

	fmt.Println(clusterPricedNodes)
}
