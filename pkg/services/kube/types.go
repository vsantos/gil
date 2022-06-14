package kube

import (
	"gil/calculator"
	"gil/pricer"

	"k8s.io/client-go/kubernetes"
)

type ClusterInterface interface {
	Prices(pricer.ProviderNodes) (ClusterPriceInterface, error)
}

type ClusterPriceInterface interface {
	List(namespace string, labelSelector string) ([]ClusterPrice, error)
}

type KubeClientConf struct {
	// current cluster-context to be used
	Context string
	// current kube's .config path
	Path string
}

type KubeConf struct {
	Client kubernetes.Interface
	Region string
}

type ClusterNode struct {
	Type           string
	Region         string
	Resources      pricer.NodeResources
	CalculatedCost calculator.NodePrice
}

type ClusterPodPrice struct {
	Name   string               `json:"name,omitempty"`
	Prices calculator.NodePrice `json:"prices,omitempty"`
}

type ClusterDeploymentPrice struct {
	Name     string               `json:"name,omitempty"`
	Replicas int32                `json:"replicas,omitempty"`
	Prices   calculator.NodePrice `json:"prices,omitempty"`
	Pods     []ClusterPodPrice    `json:"pods,omitempty"`
}

type ClusterPrice struct {
	Selector        string                 `json:"selector,omitempty"`
	RequestedMemory int64                  `json:"requested_memory,omitempty"`
	RequestedCPUMil float32                `json:"requested_cpu_mil,omitempty"`
	Deployment      ClusterDeploymentPrice `json:"deployment,omitempty"`
}
