package config

import (
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
		kubernetesConfig, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
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
