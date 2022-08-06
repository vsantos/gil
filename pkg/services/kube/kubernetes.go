package kube

import (
	"context"
	"errors"
	"fmt"
	"gil/calculator"
	"gil/pricer"
	"math"
	"regexp"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type ClusterPriceConf struct {
	PricedNodes        pricer.ProviderNodes
	ClusterPricedNodes map[string]ClusterNode
	Client             kubernetes.Interface
}

func (k *KubeConf) Prices(p pricer.ProviderNodes) (ClusterPriceInterface, error) {
	// Get prices for all nodes within a specific cluster
	nodes, err := k.GetNodes(context.Background(), k.Client)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, nil
	}

	// Endup with a list of all instance types used within a specific cluster
	log.Debug("getting prices for nodes deployed in region: ", k.Region)
	cns := make(map[string]ClusterNode)
	for _, node := range nodes {
		hostType := node.Labels["node.kubernetes.io/instance-type"]

		cns[node.Name] = ClusterNode{
			Type:           hostType,
			Region:         k.Region,
			Resources:      p[hostType].Resources,
			CalculatedCost: calculator.CalculateNodePrice(p[hostType].Cost.RegionalCost.Value[k.Region]),
		}
	}

	if len(cns) == 0 {
		return nil, errors.New("non empty cluster nodes")
	}

	// Based on every instance type within a specific cluster, get it's general price
	// k.ClusterPricedNodes = cns
	// var c ClusterPodsInterface
	c := &ClusterPriceConf{
		PricedNodes:        p,
		ClusterPricedNodes: cns,
		Client:             k.Client,
	}
	return c, nil
}

// List will return all estimated costs for deployments from a given namespace and labelselector
func (k *ClusterPriceConf) List(namespace string, labelSelector string) ([]ClusterPrice, error) {
	var clusterPrices []ClusterPrice
	var depPrices calculator.NodePrice
	var podPrices []ClusterPodPrice

	// Now we can get all deployments
	deployments, err := k.GetDeployments(context.Background(), namespace, labelSelector)
	if err != nil {
		return []ClusterPrice{}, err
	}

	// In case of empty deployments, we assume that the selector didn't fetch any resource
	if len(deployments) == 0 {
		return []ClusterPrice{}, nil
	}

	for _, deployment := range deployments {
		// clusterPrices = []ClusterPrice{}
		depPrices = calculator.NodePrice{}
		if *deployment.Spec.Replicas != 0 {
			log.Debug("fetching info for deployment: ", deployment.Name)
			rc, err := GetCPURequest(deployment)
			if err != nil {
				return []ClusterPrice{}, err
			}

			log.Debug("fetched requested CPU: ", rc)
			rm, err := GetMemoryRequest(deployment)
			if err != nil {
				return []ClusterPrice{}, err
			}

			log.Debug("fetched requested Memory: ", rm)
			pods, err := k.GetPods(context.Background(), namespace, deployment.Name, labelSelector)
			if err != nil {
				return []ClusterPrice{}, err
			}

			log.Debug("Associated pods num: ", len(pods))
			log.Debug(k.ClusterPricedNodes)

			podPrices = []ClusterPodPrice{}
			for _, pod := range pods {
				log.Debug("fetching info for pod: ", pod.Name)
				rPrices, err := ReturnPodPrice(*deployment.Spec.Replicas, rc, rm, pod.Spec.NodeName, k.ClusterPricedNodes)
				if err != nil {
					return []ClusterPrice{}, err
				}
				podPrices = append(podPrices, ClusterPodPrice{Name: pod.Name, Prices: rPrices})

				depPrices.Hourly += rPrices.Hourly
				depPrices.Daily += rPrices.Daily
				depPrices.Weekly += rPrices.Weekly
				depPrices.Monthly += rPrices.Monthly
				depPrices.Yearly += rPrices.Yearly

				log.Debug("fetched individual pod prices:", rPrices)
			}

			clusterPrices = append(clusterPrices, ClusterPrice{
				Selector:        labelSelector,
				RequestedMemory: rm,
				RequestedCPUMil: rc,
				Deployment: ClusterDeploymentPrice{
					Name:     deployment.Name,
					Replicas: *deployment.Spec.Replicas,
					Prices:   depPrices,
					Pods:     podPrices,
				},
			})
		}
	}

	return clusterPrices, nil
}

// ReturnPodPrice will return the individual pod price based on the highest cost for either CPU or Memory request.
// It will compare the percentage of request from a scheduled node type and gets it's estimated cost.
func ReturnPodPrice(replicas int32, podRequestCPUMil float32, podRequestMemGB int64, scheduledNode string, nodes map[string]ClusterNode) (calculator.NodePrice, error) {
	if nodes[scheduledNode].Resources.VCPU == 0 || nodes[scheduledNode].Resources.MemoryGB == 0 {
		return calculator.NodePrice{}, fmt.Errorf("empty VCPU and/org Memory attributes for node %s", scheduledNode)
	}

	memUsagePercentRounded := CalculateMemPercentageOfUsage(float32(podRequestMemGB), nodes[scheduledNode].Resources.MemoryGB)
	cpuUsagePercentRounded := CalculateCPUPercentageOfUsage(podRequestCPUMil, nodes[scheduledNode].Resources.VCPU)

	// calculate price based on which resource is more consumed in %
	var highestPercent float32
	if memUsagePercentRounded > cpuUsagePercentRounded {
		log.Debug("Memory is more consumed than CPU, it will be used to calculate the pricing")
		highestPercent = memUsagePercentRounded
	}

	if cpuUsagePercentRounded > memUsagePercentRounded {
		log.Debug("CPU is more consumed than memory, it will be used to calculate the pricing")
		highestPercent = cpuUsagePercentRounded
	}

	c := CalculatePodPriceByUsage(highestPercent, nodes[scheduledNode].CalculatedCost)
	return c, nil
}

func PercentageChange(percent float32, total float32) float32 {
	return ((percent / 100) * total)
}

func CalculatePodPriceByUsage(memUsagePercentRounded float32, nodePrice calculator.NodePrice) calculator.NodePrice {
	hourly := float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Hourly)))
	daily := float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Daily)))
	weekly := float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Weekly)))
	monthly := float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Monthly)))
	yearly := float64(PercentageChange(memUsagePercentRounded, float32(nodePrice.Yearly)))

	return calculator.NodePrice{
		Hourly:  calculator.RoundFloat(hourly, 4),
		Daily:   calculator.RoundFloat(daily, 4),
		Weekly:  calculator.RoundFloat(weekly, 4),
		Monthly: calculator.RoundFloat(monthly, 4),
		Yearly:  calculator.RoundFloat(yearly, 4),
	}
}

// CalculateCPUPercentageOfUsage
func CalculateCPUPercentageOfUsage(cpuPodRequest float32, cpuNode int) float32 {
	var p float32
	p = (cpuPodRequest / float32(cpuNode)) * 100
	return p
}

func CalculateMemPercentageOfUsage(memPodRequestBytes float32, memoryNodeGB float32) float32 {
	var memPodRequestMB float32
	var bytes float32
	bytes = 1024

	memPodRequestMB = ((memPodRequestBytes / bytes) / bytes)
	memoryNodeMB := (memoryNodeGB * 1024)
	memUsagePercent := (memPodRequestMB / memoryNodeMB) * float32(100)
	memUsagePercentRounded := math.Round(float64(memUsagePercent*100)) / 100

	return float32(memUsagePercentRounded)
}

// GetDeployments will return a list of deployments based on labelSelector
func (k *ClusterPriceConf) GetDeployments(ctx context.Context, namespace string, selector string) ([]appsv1.Deployment, error) {
	list, err := k.Client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetPods will return a list of pods based on labelSelector and filtered by parent deployment
func (k *ClusterPriceConf) GetPods(ctx context.Context, namespace string, deployment string, selector string) ([]corev1.Pod, error) {
	var filteredPods []corev1.Pod

	list, err := k.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	rQuery := fmt.Sprintf("^%s.+", deployment)
	for _, pod := range list.Items {
		isChild, _ := regexp.MatchString(rQuery, pod.Name)
		if isChild {
			filteredPods = append(filteredPods, pod)
		}
	}

	return filteredPods, nil
}

// GetCPURequest will return the CPU request from a given .Deployment object
func GetCPURequest(d appsv1.Deployment) (requested float32, err error) {
	var cpu int64
	// var cpuIsOK bool
	// var unscaledOk bool

	isCPUZero := d.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().IsZero()
	if isCPUZero {
		log.Debug("could not find CPU for: ", d.Spec.Template.Spec.Containers[0].Name)
		return 0, nil
	}

	if !isCPUZero {
		cpu = d.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().ToDec().MilliValue()
	}

	// return CPU in milicore
	var cpuMil float32
	cpuMil = float32(cpu) / 1000
	return cpuMil, nil
}

// GetMemoryRequest will return the CPU request from a given .Deployment object
func GetMemoryRequest(d appsv1.Deployment) (requested int64, err error) {
	var memory int64
	var memoryIsOK bool
	var unscaledOk bool

	isMemoryZero := d.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().IsZero()
	if isMemoryZero {
		log.Debug("could not find Memory for: ", d.Spec.Template.Spec.Containers[0].Name)
		return 0, nil
	}

	if !isMemoryZero {
		memory, memoryIsOK = d.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().AsInt64()
		if !memoryIsOK {
			memory, unscaledOk = d.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().AsDec().Unscaled()
			if !unscaledOk {
				log.Error("could not get unscaled metrics for ", d.Spec.Template.Spec.Containers[0].Name)
				return 0, nil
			}

		}
	}

	return memory, nil
}

// GetNodes will return a list of nodes from a given cluster
func (k *KubeConf) GetNodes(ctx context.Context, c kubernetes.Interface) ([]corev1.Node, error) {
	list, err := c.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []corev1.Node{}, err
	}
	return list.Items, nil
}
