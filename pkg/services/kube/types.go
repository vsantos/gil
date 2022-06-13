package kube

import (
	"gil/calculator"
	"gil/pricer"

	"k8s.io/client-go/kubernetes"
)

type ClusterInterface interface {
	Prices(pricer.ProviderNodes) ClusterPriceInterface
}

type ClusterPriceInterface interface {
	List(namespace string, labelSelector string) ([]ClusterNodePrice, error)
}

type KubeClientConf struct {
	// current cluster-context to be used
	Context string
	// current kube's .config path
	Path string
}

type KubeConf struct {
	Client *kubernetes.Clientset
}

type ClusterNode struct {
	// Host           string
	Type           string
	Region         string
	Resources      pricer.NodeResources
	CalculatedCost calculator.NodePrice
}
type ClusterNodePrice struct {
	Kind             string
	Name             string
	Replicas         int32
	Selector         string
	RequestedMemory  int64
	RequestedCPUMil  float32
	PricedPod        []calculator.NodePrice
	PricedDeployment calculator.NodePrice
}
