// Copyright 2022 Linkall Inc.
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

package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/controller"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/handlers"
	"k8s.io/klog/v2"
)

func main() {
	//address options
	addr := flag.String("addr", ":8089", "listen address")
	// namespace := flag.String("namespace", "vanus", "listen on namespace")
	k8scfg := flag.String("kubeconfig", "", "kubeconfig file path")
	basepath := flag.String("baseurl", "/api/v1", "base url prefix")
	klog.InitFlags(flag.CommandLine)

	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stdout)
	kubeconfig := controller.GetKubeConfigFromEnv()
	if *k8scfg != "" {
		kubeconfig = *k8scfg
	}
	klog.Infof("create kubernetes client, kubeconfig path: %v", kubeconfig)
	config, err := controller.GetInClusterOrKubeConfig(kubeconfig)
	if err != nil {
		panic(err)
	}
	control := controller.New(config)
	bpath := filepath.Clean(*basepath)
	klog.Infof("baseurl is: %v", bpath)
	//api init, include wrap handler
	a, err := handlers.NewApi(logger, control, bpath)
	if err != nil {
		panic(err)
	}

	engine := gin.Default()
	engine.NoRoute(func(c *gin.Context) {
		a.Handler().ServeHTTP(c.Writer, c.Request)
	})
	engine.NoMethod(func(c *gin.Context) {
		a.Handler().ServeHTTP(c.Writer, c.Request)
	})

	err = engine.Run(*addr)
	if err != nil {
		panic(err)
	}
}
