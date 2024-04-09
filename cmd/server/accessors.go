package main

import (
	"fmt"

	klog "k8s.io/klog/v2"
)

func NamespaceCachedListGet(locked bool) []string {
	self := "NamespaceCachedListGet"
	klog.Infof("%s: entry", self)
	if !locked {
		NamespacesLock.Lock()
		defer NamespacesLock.Unlock()
		locked = true
	}
	keys := make([]string, 0, len(Namespaces))
	for k := range Namespaces {
		keys = append(keys, k)
	}
	return keys
}

func NamespaceCachedExists(locked bool, nsName string) bool {
	self := "NamespaceCachedExists"
	klog.Infof("%s: entry", self)
	if !locked {
		NamespacesLock.Lock()
		defer NamespacesLock.Unlock()
		locked = true
	}
	_, ok := Namespaces[nsName]
	return ok
}

func DeploymentCachedExists(locked bool, nsName string, dName string) bool {
	self := "DeploymentCachedExists"
	klog.Infof("%s: entry", self)
	if !locked {
		NamespacesLock.Lock()
		defer NamespacesLock.Unlock()
		locked = true
	}
	_, ok := Namespaces[nsName]
	if !ok {
		return false
	}
	_, ok = Namespaces[nsName].Deployments[dName]
	return ok
}

func DeploymentCachedListGet(locked bool, nsName string) ([]DeploymentItem, error) {
	self := "DeploymentCachedListGet"
	klog.Infof("%s: entry", self)
	if !locked {
		NamespacesLock.Lock()
		defer NamespacesLock.Unlock()
		locked = true
	}
	if !NamespaceCachedExists(locked, nsName) {
		return nil, fmt.Errorf("unknown %q arg value: %q", "nsName", nsName)
	}
	namespace := Namespaces[nsName]
	deploymentList := make([]DeploymentItem, 0)
	if len(namespace.Deployments) != 0 {
		for _, d := range namespace.Deployments {
			deploymentList = append(deploymentList, DeploymentItem{d.Name, d.Replicas})
		}
	}
	return deploymentList, nil
}

func DeploymentCachedListAllGet(locked bool) ([]NamespaceListItem, error) {
	self := "DeploymentCachedListAllGet"
	klog.Infof("%s: entry", self)
	if !locked {
		NamespacesLock.Lock()
		defer NamespacesLock.Unlock()
		locked = true
	}
	namespaceList := make([]NamespaceListItem, 0)
	for _, namespace := range Namespaces {
		if len(namespace.Deployments) != 0 {
			deploymentList, _ := DeploymentCachedListGet(locked, namespace.Name)
			namespaceList = append(namespaceList, NamespaceListItem{namespace.Name, deploymentList})
		}
	}
	return namespaceList, nil
}

func ReplicasCachedListGet(locked bool, nsName string, dName string) (DeploymentItem, error) {
	self := "ReplicasCachedListGet"
	klog.Infof("%s: entry", self)
	if !locked {
		NamespacesLock.Lock()
		defer NamespacesLock.Unlock()
		locked = true
	}
	if !NamespaceCachedExists(locked, nsName) {
		return DeploymentItem{}, fmt.Errorf("unknown %q arg value: %q", "nsName", nsName)
	}
	if !DeploymentCachedExists(locked, nsName, dName) {
		return DeploymentItem{}, fmt.Errorf("unknown %q arg value: %q", "dName", dName)
	}
	return *Namespaces[nsName].Deployments[dName], nil
}

func ReplicasCachedListAllGet(locked bool, nsName string) ([]DeploymentItem, error) {
	self := "ReplicasCachedListAllGet"
	klog.Infof("%s: entry", self)
	if !locked {
		NamespacesLock.Lock()
		defer NamespacesLock.Unlock()
		locked = true
	}
	deploymentList := make([]DeploymentItem, 0)
	for _, deployment := range Namespaces[nsName].Deployments {
		deploymentList = append(deploymentList, *deployment)
	}

	return deploymentList, nil
}
