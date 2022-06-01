package main

import (
	"context"
	"fmt"
	"gil/calculator"
	"gil/price"
	"gil/services"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var instanceTypes []string

const REGION = "sa-east-1"

func main() {
	// List price for every AWS instance
	var awsPrices price.Pricer
	awsPrices = &price.AwsPricing{}
	priced := awsPrices.List()

	c, err := services.NewKubeClient("", "")
	if err != nil {
		log.Panic(err)
	}

	// Get prices for all nodes within a specific cluster
	nodes, err := GetNodes(c, context.TODO())
	if err != nil {
		fmt.Println(err)
	}

	// Endup with a list of all instance types used within a specific cluster
	var cn []services.ClusterNode
	for _, node := range nodes {
		hostType := node.Labels["node.kubernetes.io/instance-type"]
		instanceTypes = append(instanceTypes, hostType)
	}

	// Based on every instance type within a specific cluster, get it's general price
	instanceTypes = unique(instanceTypes)
	for _, instance := range instanceTypes {
		cn = append(cn, services.ClusterNode{
			Type:           instance,
			Region:         REGION,
			Resources:      priced[instance].Resources,
			CalculatedCost: calculator.CalculateNodePrice(priced[instance].Cost.RegionalCost.Value[REGION]),
		})
	}
	fmt.Println(cn)

	// Now we can get all deployments
	items, err := GetDeployments(c, context.TODO(), "kube-system")
	if err != nil {
		fmt.Println(err)
	} else {
		// var mem, cpu int64
		for _, item := range items {
			fmt.Println(item.Name)
			fmt.Println(item.Spec.Template.Spec.NodeName)
			if item.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().IsZero() {
				fmt.Println("could not find CPU for: ", item.Spec.Template.Spec.Containers[0].Name)
			} else {
				cpu, isInt := item.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().AsInt64()
				fmt.Println(isInt)
				if !isInt {
					cpu := item.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().ToDec().Value()
					fmt.Println("1 cpu: ", cpu)
				}
				fmt.Println(cpu)
			}

			if item.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().IsZero() {
				fmt.Println("could not find Memory for: ", item.Spec.Template.Spec.Containers[0].Name)
			} else {
				mem, isInt := item.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().AsInt64()
				if !isInt {
					mem = item.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().ToDec().Value()
				} else {
					fmt.Println("mem2: ", mem)
				}
			}

		}
	}

}

func GetDeployments(clientset *kubernetes.Clientset, ctx context.Context,
	namespace string) ([]appsv1.Deployment, error) {

	list, err := clientset.AppsV1().Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetNodes(clientset *kubernetes.Clientset, ctx context.Context) ([]corev1.Node, error) {
	list, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func unique(s []string) []string {
	inResult := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}
