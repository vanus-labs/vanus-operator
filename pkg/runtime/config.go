// Copyright 2023 Linkall Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	clientset "github.com/vanus-labs/vanus-operator/pkg/runtime/client/clientset/versioned"
)

// Config is the controller conf
type Config struct {
	KubeConfigFile     string
	KubeRestConfig     *rest.Config
	KubeFactoryClient  kubernetes.Interface
	VanusFactoryClient clientset.Interface
}

// ParseFlags parses cmd args then init kubeclient and conf
// TODO: validate configuration
func NewConfig() (*Config, error) {
	config := &Config{}
	if err := config.initKubeFactoryClient(); err != nil {
		return nil, err
	}
	return config, nil
}

func (config *Config) initKubeFactoryClient() error {
	var cfg *rest.Config
	var err error
	config.KubeConfigFile = GetKubeConfigFromEnv()
	cfg, err = GetInClusterOrKubeConfig(config.KubeConfigFile)
	if err != nil {
		klog.Errorf("failed to build kubeconfig %v", err)
		return err
	}
	cfg.QPS = 1000
	cfg.Burst = 2000

	config.KubeRestConfig = cfg

	VanusClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("init vanus client failed %v", err)
		return err
	}
	config.VanusFactoryClient = VanusClient

	cfg.ContentType = "application/vnd.kubernetes.protobuf"
	cfg.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("init kubernetes client failed %v", err)
		return err
	}
	config.KubeFactoryClient = kubeClient
	return nil
}

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
