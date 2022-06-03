package services

import (
	"context"
	"errors"
	"fmt"
	"gil/calculator"
	"gil/pricer"
	"math"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type ClusterPriceConf struct {
	PricedNodes        pricer.ProviderNodes
	ClusterPricedNodes map[string]ClusterNode
	Client             *kubernetes.Clientset
}

func (k *KubeConf) Pods(p pricer.ProviderNodes) ClusterPodsInterface {
	var c ClusterPodsInterface
	c = &ClusterPriceConf{
		PricedNodes: p,
		Client:      k.Client,
	}

	fmt.Println("price prices")
	return c
}

func (k *ClusterPriceConf) Prices() ClusterPriceInterface {
	fmt.Println("listing cluster nodes")
	// Get prices for all nodes within a specific cluster
	nodes, err := GetNodes(k.Client, context.TODO())
	if err != nil {
		return nil
	}

	if len(nodes) == 0 {
		return nil
	}

	// Endup with a list of all instance types used within a specific cluster
	// var cn []services.ClusterNode
	// var cns map[string]services.ClusterNode
	cns := make(map[string]ClusterNode)
	for _, node := range nodes {
		hostType := node.Labels["node.kubernetes.io/instance-type"]

		cns[node.Name] = ClusterNode{
			Type:           hostType,
			Region:         "sa-east-1",
			Resources:      k.PricedNodes[hostType].Resources,
			CalculatedCost: calculator.CalculateNodePrice(k.PricedNodes[hostType].Cost.RegionalCost.Value["sa-east-1"]),
		}
	}

	if len(cns) == 0 {
		return nil
	}

	// Based on every instance type within a specific cluster, get it's general price
	k.ClusterPricedNodes = cns
	return k
}

func (k *ClusterPriceConf) List(namespace string, labelSelector string) (calculator.NodePrice, error) {
	var podSumPrices calculator.NodePrice

	// Now we can get all deployments
	deployments, err := k.GetDeployments(context.TODO(), namespace, labelSelector)
	if err != nil {
		return calculator.NodePrice{}, err
	}

	for _, deployment := range deployments {
		if *deployment.Spec.Replicas != 0 {
			fmt.Println("Deployment: ", deployment.Name)
			rc, err := GetCPURequest(deployment)
			if err != nil {
				return calculator.NodePrice{}, err
			}

			fmt.Println("Requests CPU: ", rc)

			rm, err := GetMemoryRequest(deployment)
			if err != nil {
				return calculator.NodePrice{}, err
			}

			fmt.Println("Requests Memory: ", rm)

			pods, err := k.GetPods(context.TODO(), namespace, labelSelector)
			if err != nil {
				return calculator.NodePrice{}, err
			}

			fmt.Println("Associated pods num: ", len(pods))
			fmt.Println(k.ClusterPricedNodes)
			for _, pod := range pods {
				rPrices, err := ReturnPodPrice(*deployment.Spec.Replicas, rc, rm, pod.Spec.NodeName, k.ClusterPricedNodes)
				if err != nil {
					return calculator.NodePrice{}, err
				}

				podSumPrices.Hourly += rPrices.Hourly
				podSumPrices.Daily += rPrices.Daily
				podSumPrices.Weekly += rPrices.Weekly
				podSumPrices.Monthly += rPrices.Monthly
				podSumPrices.Yearly += rPrices.Yearly

				fmt.Println(pod.Name)
				fmt.Println("cost per pod: ", rPrices)
			}

			fmt.Println("cost per deployment: ", podSumPrices)
		}

	}

	return podSumPrices, nil
}

func ReturnPodPrice(replicas int32, podRequestCPU int64, podRequestMem int64, scheduledNode string, nodes map[string]ClusterNode) (calculator.NodePrice, error) {
	if nodes[scheduledNode].Resources.VCPU == 0 || nodes[scheduledNode].Resources.MemoryGB == 0 {
		return calculator.NodePrice{}, errors.New(fmt.Sprintf("empty VCPU and/org Memory attributes for node %s", scheduledNode))
	}

	memUsagePercentRounded := CalculatePercentageOfUsage(float32(podRequestMem), nodes[scheduledNode].Resources.MemoryGB)

	c := CalculatePodPriceByUsage(memUsagePercentRounded, nodes[scheduledNode].CalculatedCost)
	fmt.Println("original cost:", nodes[scheduledNode].CalculatedCost.Hourly)
	fmt.Println("percent of cost:", c.Hourly)

	return c, nil
}

func PercentageChange(percent float32, total float32) float32 {
	return ((percent / 100) * total)
}

func CalculatePodPriceByUsage(memUsagePercentRounded float32, nodePrice calculator.NodePrice) calculator.NodePrice {
	return calculator.NodePrice{
		Hourly:  float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Hourly))),
		Daily:   float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Daily))),
		Weekly:  float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Weekly))),
		Monthly: float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Monthly))),
		Yearly:  float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Yearly))),
	}
}

func CalculatePercentageOfUsage(memPodRequestBytes float32, memoryNodeGB float32) float32 {
	var memPodRequestMB float32
	var bytes float32
	bytes = 1024

	memPodRequestMB = ((memPodRequestBytes / bytes) / bytes)
	memoryNodeMB := (memoryNodeGB * 1024)
	memUsagePercent := (memPodRequestMB / memoryNodeMB) * float32(100)
	memUsagePercentRounded := math.Round(float64(memUsagePercent*100)) / 100

	fmt.Println(memPodRequestMB)
	fmt.Println(memoryNodeMB)
	fmt.Println("% naked:", memUsagePercent)
	fmt.Println("% round:", memUsagePercentRounded)

	return float32(memUsagePercentRounded)
}

func (k *ClusterPriceConf) GetDeployments(ctx context.Context, namespace string, selector string) ([]appsv1.Deployment, error) {

	list, err := k.Client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (k *ClusterPriceConf) GetPods(ctx context.Context, namespace string, selector string) ([]corev1.Pod, error) {

	list, err := k.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetCPURequest(d appsv1.Deployment) (requested int64, err error) {
	var cpu int64
	var cpuIsOK bool
	var unscaledOk bool

	isCPUZero := d.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().IsZero()
	if isCPUZero {
		fmt.Println("could not find CPU for: ", d.Spec.Template.Spec.Containers[0].Name)
		return 0, nil
	}

	if !isCPUZero {
		cpu, cpuIsOK = d.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().AsInt64()
		if !cpuIsOK {
			cpu, unscaledOk = d.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().AsDec().Unscaled()
			if !unscaledOk {
				fmt.Println("could not get unscaled metrics for ", d.Spec.Template.Spec.Containers[0].Name)
				return 0, nil
			}

		}
	}

	return cpu, nil
}

func GetMemoryRequest(d appsv1.Deployment) (requested int64, err error) {
	var memory int64
	var memoryIsOK bool
	var unscaledOk bool

	isMemoryZero := d.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().IsZero()
	if isMemoryZero {
		fmt.Println("could not find Memory for: ", d.Spec.Template.Spec.Containers[0].Name)
		return 0, nil
	}

	if !isMemoryZero {
		memory, memoryIsOK = d.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().AsInt64()
		if !memoryIsOK {
			memory, unscaledOk = d.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().AsDec().Unscaled()
			if !unscaledOk {
				fmt.Println("could not get unscaled metrics for ", d.Spec.Template.Spec.Containers[0].Name)
				return 0, nil
			}

		}
	}

	return memory, nil
}

func GetNodes(clientset *kubernetes.Clientset, ctx context.Context) ([]corev1.Node, error) {
	list, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
