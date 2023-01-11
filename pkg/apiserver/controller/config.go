// Copyright 2017 EasyStack, Inc.
package controller

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func GetKubeConfigFromEnv() string {
	home := os.Getenv("HOME")
	if home != "" {
		fpath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(fpath); err != nil {
			return ""
		}
		return fpath
	}
	return ""
}

// try list authprovider
// kubeconfig
// serviceAccount
func GetInClusterOrKubeConfig(kubeconfig string) (config *rest.Config, rerr error) {
	config, rerr = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if rerr != nil {
		klog.Errorf("auth from kubeconfig failed:%v", rerr)
		return nil, rerr
	}
	return config, nil
}
