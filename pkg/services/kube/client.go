package kube

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ToRawKubeConfigLoader returns a ClientConfig with overrided attributes such as 'context'
func toRawKubeConfigLoader(kubeContext string, kubeConfigPath string) clientcmd.ClientConfig {

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// if you want to change the loading rules (which files in which order), you can do so here
	loadingRules.ExplicitPath = kubeConfigPath

	// if you want to change override values or bind them to flags, there are methods to help you
	configOverrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
	}

	if kubeContext != "" {
		configOverrides.CurrentContext = kubeContext
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig
}

func (k *KubeClientConf) NewKubeClient() (*kubernetes.Clientset, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		return nil, err
	}

	// alow custom kubeConfigPath to be passed
	if k.Path == "" {
		k.Path = filepath.Join(userHomeDir, ".kube", "config")
	}

	kubeConfig, err := toRawKubeConfigLoader(k.Context, k.Path).ClientConfig()
	if err != nil {
		fmt.Printf("error getting Kubernetes config: %v\n", err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		fmt.Printf("error getting Kubernetes clientset: %v\n", err)
		return nil, err
	}

	return clientset, nil
}
