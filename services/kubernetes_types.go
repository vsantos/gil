package services

import (
	"gil/calculator"
	"gil/price"

	"k8s.io/client-go/kubernetes"
)

type KubeConf struct {
	Client *kubernetes.Clientset
}
type ClusterNode struct {
	Type           string
	Region         string
	Resources      price.NodeResources
	CalculatedCost calculator.NodePrice
}

type Noder interface {
	Get() []ClusterNode
}
