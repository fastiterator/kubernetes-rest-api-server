package main

import (
	"fmt"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
)

var (
	InformersSavedApp *AppX
)

type NamespaceLoggingController struct {
	informerFactory   informers.SharedInformerFactory
	namespaceInformer coreinformers.NamespaceInformer
}

type DeploymentLoggingController struct {
	informerFactory    informers.SharedInformerFactory
	deploymentInformer appsinformers.DeploymentInformer
}

type StringList []string
type MapStringList map[string]StringList

type DeploymentItem struct {
	Name     string `json:"deployment"`
	Replicas int    `json:"replica_count"`
}
type DeploymentMap map[string]*DeploymentItem

type NamespaceItem struct {
	Name        string        `json:"namespace"`
	Deployments DeploymentMap `json:"deployments"`
}
type NamespaceMap map[string]*NamespaceItem
type NamespaceListItem struct {
	Name        string           `json:"namespace"`
	Deployments []DeploymentItem `json:"deployments"`
}

var (
	Namespaces     NamespaceMap = make(NamespaceMap)
	NamespacesLock sync.Mutex
)

func (c *DeploymentLoggingController) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.deploymentInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}

func (c *DeploymentLoggingController) deploymentAdd(obj interface{}) {
	NamespacesLock.Lock()
	defer NamespacesLock.Unlock()
	self := "deploymentAdd"
	deploymentObject := obj.(*appsv1.Deployment)
	nsName, dName, dReplicas := deploymentObject.Namespace, deploymentObject.Name, int(*deploymentObject.Spec.Replicas)
	dMap, ok := Namespaces[nsName]
	if !ok {
		klog.Errorf("%s: event refs unknown namespace: %q", self, nsName)
		return
	}
	_, ok = dMap.Deployments[dName]
	if ok {
		klog.Errorf("%s: event refs existing deployment: \"%s/%s\"", self, nsName, dName)
		return
	}
	di := new(DeploymentItem)
	di.Name, di.Replicas = dName, dReplicas
	Namespaces[nsName].Deployments[dName] = di
	klog.Infof("%s: created: \"%s/%s\"  replicas=%d", self, nsName, dName, dReplicas)
}

func (c *DeploymentLoggingController) deploymentUpdate(old, new interface{}) {
	NamespacesLock.Lock()
	defer NamespacesLock.Unlock()
	self := "deploymentUpdate"
	oldDeployment, newDeployment := old.(*appsv1.Deployment), new.(*appsv1.Deployment)
	oldReplicas, newReplicas := *oldDeployment.Spec.Replicas, *newDeployment.Spec.Replicas
	oldNS, oldName := oldDeployment.Namespace, oldDeployment.Name
	newNS, newName := newDeployment.Namespace, newDeployment.Name
	nameChange, replicasChange := oldName != newName, oldReplicas != newReplicas
	if oldNS != newNS {
		klog.Errorf("%s: event includes namespace name change: old=%#v;  new=%#v", self, oldDeployment, newDeployment)
	}
	dMap, ok := Namespaces[oldNS]
	if !ok {
		klog.Errorf("%s: event refs unknown namespace: %q", self, oldNS)
		return
	}
	_, ok = dMap.Deployments[oldName]
	if !ok {
		klog.Errorf("%s: event refs unknown orig deployment: \"%s/%s\"", self, oldNS, oldName)
		return
	}
	_, ok = dMap.Deployments[newName]
	if nameChange && ok {
		klog.Errorf("%s: event refs existing new name: \"%s/%s\" -> \"%s/%s\"", self, oldNS, oldName, oldNS, newName)
		return
	}
	if !nameChange && !replicasChange {
		klog.Errorf("%s: event is not deployment name change or replica count change. ignored.", self)
		return
	}
	if nameChange {
		Namespaces[oldNS].Deployments[newName] = Namespaces[oldNS].Deployments[oldName]
		delete(Namespaces[oldNS].Deployments, oldName)
		klog.Infof("%s: name updated: \"%s/%s\" -> \"%s/%s\"", self, oldNS, oldName, oldNS, newName)
	}
	if replicasChange {
		Namespaces[oldNS].Deployments[newName].Replicas = int(newReplicas)
		klog.Infof("%s: replica count updated: \"%s/%s\": %d -> %d", self, oldNS, newName, oldReplicas, newReplicas)
	}
}

func (c *DeploymentLoggingController) deploymentDelete(obj interface{}) {
	NamespacesLock.Lock()
	defer NamespacesLock.Unlock()
	self := "deploymentDelete"
	deployment := obj.(*appsv1.Deployment)
	nsName, name := deployment.Namespace, deployment.Name
	dMap, ok := Namespaces[nsName]
	if !ok {
		klog.Errorf("%s: event refs unknown namespace: %q", self, nsName)
		return
	}
	_, ok = dMap.Deployments[name]
	if !ok {
		klog.Errorf("%s: event refs unknown deployment: \"%s/%s\"", self, nsName, name)
		return
	}
	delete(Namespaces[nsName].Deployments, name)
	klog.Infof("%s: deleted: \"%s/%s\"", self, nsName, name)
}

func (c *NamespaceLoggingController) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.namespaceInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}

func (c *NamespaceLoggingController) namespaceAdd(obj interface{}) {
	NamespacesLock.Lock()
	defer NamespacesLock.Unlock()
	self := "namespaceAdd"
	namespaceObject := obj.(*corev1.Namespace)
	nsName := namespaceObject.Name
	if _, ok := Namespaces[nsName]; ok {
		klog.Errorf("%s: event refs existing namespace: %q", self, nsName)
		return
	}
	ni := new(NamespaceItem)
	ni.Name, ni.Deployments = nsName, make(DeploymentMap)
	Namespaces[nsName] = ni
	klog.Infof("%s: created: %q", self, nsName)
}

func (c *NamespaceLoggingController) namespaceUpdate(old, new interface{}) {
	NamespacesLock.Lock()
	defer NamespacesLock.Unlock()
	self := "namespaceUpdate"
	oldNamespace := old.(*corev1.Namespace)
	newNamespace := new.(*corev1.Namespace)
	oldName, newName := oldNamespace.Name, newNamespace.Name
	if oldName == newName {
		klog.Errorf("%s: event is not name change. ignored.", self)
		return
	}
	if _, ok := Namespaces[oldName]; !ok {
		klog.Errorf("%s: event refs unknown old namespace: %q", self, oldName)
		return
	}
	if _, ok := Namespaces[newName]; ok {
		klog.Errorf("%s: event refs existing new namespace: %q", self, newName)
		return
	}
	Namespaces[newName] = Namespaces[oldName]
	delete(Namespaces, oldName)
	klog.Infof("%s: updated: %q -> %q", self, oldName, newName)
}

func (c *NamespaceLoggingController) namespaceDelete(obj interface{}) {
	NamespacesLock.Lock()
	defer NamespacesLock.Unlock()
	self := "namespaceDelete"
	namespaceObject := obj.(*corev1.Namespace)
	nsName := namespaceObject.Name
	if _, ok := Namespaces[nsName]; !ok {
		klog.Errorf("%s: event refs unknown namespace: %q", self, nsName)
		return
	}
	delete(Namespaces, nsName)
	klog.Infof("%s: deleted: %q", self, nsName)
}

func NewDeploymentLoggingController(informerFactory informers.SharedInformerFactory) (*DeploymentLoggingController, error) {
	deploymentInformer := informerFactory.Apps().V1().Deployments()
	c := &DeploymentLoggingController{
		informerFactory:    informerFactory,
		deploymentInformer: deploymentInformer,
	}
	_, err := deploymentInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.deploymentAdd,
			UpdateFunc: c.deploymentUpdate,
			DeleteFunc: c.deploymentDelete,
		},
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewNamespaceLoggingController(informerFactory informers.SharedInformerFactory) (*NamespaceLoggingController, error) {
	namespaceInformer := informerFactory.Core().V1().Namespaces()
	c := &NamespaceLoggingController{
		informerFactory:   informerFactory,
		namespaceInformer: namespaceInformer,
	}
	_, err := namespaceInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.namespaceAdd,
			UpdateFunc: c.namespaceUpdate,
			DeleteFunc: c.namespaceDelete,
		},
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func initInformers(App *AppX) error {
	self := "initInformers"
	factory := informers.NewSharedInformerFactory(App.Clientset, time.Hour*24)
	namespaceLoggingController, err := NewNamespaceLoggingController(factory)
	if err != nil {
		return fmt.Errorf("%s: call to %q failed: %#v", self, "clientcmd.BuildConfigFromFlags", err)
	}
	deploymentLoggingController, err := NewDeploymentLoggingController(factory)
	if err != nil {
		return fmt.Errorf("%s: call to %q failed: %#v", self, "clientcmd.BuildConfigFromFlags", err)
	}
	App.Stop = make(chan struct{})
	err = namespaceLoggingController.Run(App.Stop)
	if err != nil {
		return fmt.Errorf("%s: call to %q failed: %#v", self, "clientcmd.BuildConfigFromFlags", err)
	}
	err = deploymentLoggingController.Run(App.Stop)
	if err != nil {
		return fmt.Errorf("%s: call to %q failed: %#v", self, "clientcmd.BuildConfigFromFlags", err)
	}
	InformersSavedApp = App
	return nil
}
