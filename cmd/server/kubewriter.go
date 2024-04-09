package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ReplicasSet(App *AppX, nsName string, dName string, replicas int) error {
	self := "ReplicasSet"
	s, err := App.Clientset.AppsV1().Deployments(nsName).GetScale(context.TODO(), dName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("%s: call to %q failed: %#v",
			self, "(clientset).AppsV1().Deployments().GetScale()", err)
	}
	sc := *s
	s, err = App.Clientset.AppsV1().Deployments(nsName).UpdateScale(context.TODO(), dName, &sc, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("%s: error while scaling deployment: \"%s/%s\" to replicas=%d: %#v",
			self, nsName, dName, replicas, err)
	}
	return nil
}
