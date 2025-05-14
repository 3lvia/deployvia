package config

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func configureKubernetesClient(local bool) (*dynamic.DynamicClient, error) {
	kubernetesConfig, err := configureKubernetesConfig(local)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, err
	}

	return dynamicClient, nil
}

func configureKubernetesConfig(local bool) (*rest.Config, error) {
	if local {
		kubeconfigPath := func() string {
			kubeconfigEnv := os.Getenv("KUBECONFIG")
			if kubeconfigEnv != "" {
				return kubeconfigEnv
			}

			return filepath.Join(homedir.HomeDir(), ".kube", "config")
		}()

		kubernetesConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}

		return kubernetesConfig, nil
	} else {
		kubernetesConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		return kubernetesConfig, nil
	}
}
