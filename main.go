package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"

	"github.com/ishtiaqhimel/crd-controller/controller"
	clientset "github.com/ishtiaqhimel/crd-controller/pkg/client/clientset/versioned"
	informers "github.com/ishtiaqhimel/crd-controller/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	log.Println("Configure KubeConfig...")

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	exampleClient, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Initialise the informer resource and here we will be using sharedinformer factory instead of simple informers
	// because in case if we need to query / watch multiple Group versions, and it’s a good practise as well
	// NewSharedInformerFactory will create a new ShareInformerFactory for "all namespaces"
	// 30*time.Second is the re-sync period to update the in-memory cache of informer //
	kubeInformationFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformationFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	// From this informerfactory we can create specific informers for every group version resource
	// that are default available in k8s environment such as Pods, deployment, etc
	// podInformer := kubeInformationFactory.Core().V1().Pods()

	ctrl := controller.NewController(kubeClient, exampleClient,
		kubeInformationFactory.Apps().V1().Deployments(),
		exampleInformationFactory.Crd().V1().Ishtiaqs())

	// creating a unbuffered channel to synchronize the update
	stopCh := make(chan struct{})
	kubeInformationFactory.Start(stopCh)
	exampleInformationFactory.Start(stopCh)

	if err = ctrl.Run(2, stopCh); err != nil {
		log.Println("Error running controller")
	}
}
