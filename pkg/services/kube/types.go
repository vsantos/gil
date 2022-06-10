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
	List(namespace string, labelSelector string) (calculator.NodePrice, error)
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
	RequestedMemory  float64
	RequestedCPU     float64
	PricedPod        calculator.NodePrice
	PricedDeployment calculator.NodePrice
}
