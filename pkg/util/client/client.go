package client

import (
	"os"
	"path/filepath"

	authv1 "github.com/KETI-Hybrid/keti-controller/apis/auth/v1"
	cloudv1 "github.com/KETI-Hybrid/keti-controller/apis/cloud/v1"
	levelv1 "github.com/KETI-Hybrid/keti-controller/apis/level/v1"
	resourcev1 "github.com/KETI-Hybrid/keti-controller/apis/resource/v1"
	keti "github.com/KETI-Hybrid/keti-controller/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	kubeClient kubernetes.Interface
	ketiClient keti.Interface
)

func init() {
	kubeClient, _ = NewClient()
}

func GetClient() kubernetes.Interface {
	return kubeClient
}

// NewClient connects to an API server
func NewClient() (kubernetes.Interface, error) {
	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig == "" {
		kubeConfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Infoln("InClusterConfig failed", err.Error())
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			klog.Errorln("BuildFromFlags failed", err.Error())
			return nil, err
		}
	}
	client, err := kubernetes.NewForConfig(config)
	return client, err
}

func NewKETIClient() (keti.Interface, error) {
	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig == "" {
		kubeConfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	err := authv1.AddToScheme(scheme.Scheme)
	if err != nil {
		klog.Errorln(err)
	}
	err = cloudv1.AddToScheme(scheme.Scheme)
	if err != nil {
		klog.Errorln(err)
	}
	err = resourcev1.AddToScheme(scheme.Scheme)
	if err != nil {
		klog.Errorln(err)
	}
	err = levelv1.AddToScheme(scheme.Scheme)
	if err != nil {
		klog.Errorln(err)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Infoln("InClusterConfig failed", err.Error())
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			klog.Errorln("BuildFromFlags failed", err.Error())
			return nil, err
		}
	}
	return keti.NewForConfig(config)
}
