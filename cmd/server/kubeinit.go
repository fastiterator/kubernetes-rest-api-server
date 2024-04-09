package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func initKubeconfig(App *AppX) error {
	self := "initKubeconfig"
	homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("%s: call to %q failed: %#v", self, "os.UserHomeDir", err)
	}
	flag.StringVar(&App.Kubeconfig, "kubeconfig", filepath.Join(homedir, ".kube", "config"),
		"absolute path to the kubeconfig file")
	flag.StringVar(&App.Port, "port", "8088", "server port")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", App.Kubeconfig)
	if err != nil {
		return fmt.Errorf("%s: call to %q failed: %#v", self, "clientcmd.BuildConfigFromFlags", err)
	}
	App.Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("%s: call to %q failed: %#v", self, "kubernetes.NewForConfig", err)
	}
	return nil
}
