package services

import (
	"gil/calculator"
	"gil/pricer"

	"k8s.io/client-go/kubernetes"
)

type ClusterInterface interface {
	Pods(pricer.ProviderNodes) ClusterPodsInterface
	// Deployments() ClusterDeploymentsInterface
	// Pods() ClusterPodsInterface
}

type ClusterPodsInterface interface {
	Prices() ClusterPriceInterface
}

type ClusterPriceInterface interface {
	List(namespace string, labelSelector string) (calculator.NodePrice, error)
}

// type ProviderInterface interface {
// 	Nodes() ProviderNodesInterface
// }

// type ProviderNodesInterface interface {
// 	Prices() PriceInterface
// }

// type PriceInterface interface {
// 	List() ProviderNodes
// }

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
