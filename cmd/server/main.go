package main

import (
	"flag"
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/examples/util/require"
	klog "k8s.io/klog/v2"
)

type AppX struct {
	Kubeconfig string
	Clientset  *kubernetes.Clientset
	Port       string
	Mux        *http.ServeMux
	Stop       chan struct{}
}

var (
	App AppX
)

func main() {
	klog.InitFlags(nil)
	require.NoError(flag.Set("logtostderr", "false"))
	require.NoError(flag.Set("log_file", "/users/nobody/log.log"))
	flag.Parse()
	defer klog.Flush()
	klog.Info("nice to meet you")
	err := initKubeconfig(&App)
	if err != nil {
		klog.Fatal(err)
	}
	err = initInformers(&App)
	if err != nil {
		klog.Fatal(err)
	}
	defer close(App.Stop)
	err = initHttp(&App)
	if err != nil {
		klog.Fatal(err)
	}
	err = http.ListenAndServe(":"+App.Port, App.Mux)
	if err != nil {
		klog.Fatal(err)
	}
}
