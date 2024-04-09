package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	klog "k8s.io/klog/v2"
)

var (
	HttpSavedApp *AppX

	reLivez                   = regexp.MustCompile(`^\/livez[\/]?$`)
	reReadyz                  = regexp.MustCompile(`^\/readyz[\/]?$`)
	reNamespaces              = regexp.MustCompile(`^\/namespaces[\/]?$`)
	reNamespaceOneDeployments = regexp.MustCompile(`^\/namespaces\/([-a-z0-9]+)\/deployments[\/]?$`)
	reNamespaceAllDeployments = regexp.MustCompile(`^\/namespaces\/ANY\/deployments[\/]?$`)
	reDeploymentOneReplicas   = regexp.MustCompile(`^\/namespaces\/([-a-z0-9]+)\/deployments[\/]([-a-z0-9]+)\/replica_count[\/]?$`)
	reDeploymentAllReplicas   = regexp.MustCompile(`^\/namespaces\/([-a-z0-9]+)\/deployments[\/]ANY\/replica_count[\/]?$`)
	reDeploymentSetReplicas   = regexp.MustCompile(`^\/namespaces\/([-a-z0-9]+)\/deployments[\/]([-a-z0-9]+)\/replica_count\/(\d+)[\/]?$`)
)

type handler struct {
	// needed solely to contain the serveHTTP method
}
type NamespaceDeployments struct {
	Namespace   string   `json:"namespace"`
	Deployments []string `json:"deployments"`
}
type NamespaceDeploymentReplica struct {
	Namespace  string `json:"namespace"`
	Deployment string `json:"deployment"`
	Replicas   int    `json:"replica_count"`
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self := "ServeHTTP"
	klog.Infof("%s: entry", self)
	klog.Infof("%s: r.URL.Path=%q", self, r.URL.Path)
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodGet && reLivez.MatchString(r.URL.Path):
		serveLiveness(w, r)
		return
	case r.Method == http.MethodGet && reReadyz.MatchString(r.URL.Path):
		serveReadiness(w, r)
		return
	case r.Method == http.MethodGet && reNamespaces.MatchString(r.URL.Path):
		serveNamespacesGet(w, r)
		return
	case r.Method == http.MethodGet && reNamespaceOneDeployments.MatchString(r.URL.Path):
		serveDeploymentsGet(w, r)
		return
	case r.Method == http.MethodGet && reNamespaceAllDeployments.MatchString(r.URL.Path):
		serveDeploymentsAllGet(w, r)
		return
	case r.Method == http.MethodGet && reDeploymentOneReplicas.MatchString(r.URL.Path):
		serveReplicasGet(w, r)
		return
	case r.Method == http.MethodGet && reDeploymentAllReplicas.MatchString(r.URL.Path):
		serveReplicasAllGet(w, r)
		return
	case (r.Method == http.MethodGet || r.Method == http.MethodPut) && reDeploymentSetReplicas.MatchString(r.URL.Path):
		serveReplicasSet(w, r)
		return
	default:
		respondWithNotFound(w, r, "unknown url path", r.URL.Path)
		return
	}
}

func respondWithNotFound(w http.ResponseWriter, r *http.Request, msg string, elt string) {
	self := "respondWithNotFound"
	klog.Infof("%s: entry", self)
	resp := make(map[string]string)
	resp["message"], resp["element"] = msg, elt
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		klog.Fatalf("call to json.Marshal() failed: %#v", err)
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write(jsonResp)
}

func respondWithInternalServerError(w http.ResponseWriter, r *http.Request, msg string, elt string, err error) {
	self := "respondWithInternalServerError"
	klog.Infof("%s: entry", self)
	resp := make(map[string]string)
	if msg == "" {
		msg = "func call returned error"
	}
	resp["message"], resp["element"], resp["error"] = msg, elt, fmt.Sprintf("%#v", err)
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		klog.Fatalf("call to json.Marshal() failed: %#v", err)
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsonResp)
}

// Endpoint #1
func serveNamespacesGet(w http.ResponseWriter, r *http.Request) {
	self := "serveNamespacesGet"
	klog.Infof("%s: entry", self)
	namespaces := StringList(NamespaceCachedListGet(false))
	klog.Infof("%s: namespaces=%#v", self, namespaces)
	namespacesMap := make(MapStringList)
	namespacesMap["namespaces"] = namespaces
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(namespacesMap)
}

// Endpoint #2
func serveDeploymentsGet(w http.ResponseWriter, r *http.Request) {
	self := "serveDeploymentsGet"
	klog.Infof("%s: entry", self)
	matches := reNamespaceOneDeployments.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		respondWithNotFound(w, r, "invalid arg(s)", r.URL.Path)
		return
	}
	nsName := matches[1]
	klog.Infof("%s: nsName=%q", self, nsName)
	if !NamespaceCachedExists(false, nsName) {
		respondWithNotFound(w, r, "namespace not found", nsName)
		return
	}
	deploymentList, err := DeploymentCachedListGet(false, nsName)
	if err != nil {
		respondWithInternalServerError(w, r, "", "DeploymentCachedListGet", err)
		return
	}
	klog.Infof("%s: deploymentList=%#v", self, deploymentList)
	stringList := []string{}
	for _, deployment := range deploymentList {
		stringList = append(stringList, deployment.Name)
	}
	namespaceDeployments := NamespaceDeployments{Namespace: nsName, Deployments: stringList}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(namespaceDeployments)
}

// Endpoint #2A
func serveDeploymentsAllGet(w http.ResponseWriter, r *http.Request) {
	self := "serveDeploymentsAllGet"
	klog.Infof("%s: entry", self)
	namespaceList, err := DeploymentCachedListAllGet(false)
	if err != nil {
		respondWithInternalServerError(w, r, "", "DeploymentCachedListAllGet", err)
		return
	}
	klog.Infof("%s: namespaceList=%#v", self, namespaceList)
	namespaceDeployments := []NamespaceDeployments{}
	for _, namespace := range namespaceList {
		if len(namespace.Deployments) > 0 {
			stringList := []string{}
			for _, deployment := range namespace.Deployments {
				stringList = append(stringList, deployment.Name)
			}
			namespaceDeployments = append(namespaceDeployments,
				NamespaceDeployments{Namespace: namespace.Name, Deployments: stringList})
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(namespaceDeployments)
}

// Endpoint #3
func serveReplicasGet(w http.ResponseWriter, r *http.Request) {
	self := "serveReplicasGet"
	klog.Infof("%s: entry", self)
	matches := reDeploymentOneReplicas.FindStringSubmatch(r.URL.Path)
	if len(matches) < 3 {
		respondWithNotFound(w, r, "invalid arg(s)", r.URL.Path)
		return
	}
	nsName, dName := matches[1], matches[2]
	klog.Infof("%s: nsName=%q;  dName=%q", self, nsName, dName)
	if !NamespaceCachedExists(false, nsName) {
		respondWithNotFound(w, r, "namespace not found", nsName)
		return
	}
	if !DeploymentCachedExists(false, nsName, dName) {
		respondWithNotFound(w, r, "deployment not found", fmt.Sprintf("%s/%s", nsName, dName))
		return
	}
	deploymentList, err := ReplicasCachedListGet(false, nsName, dName)
	if err != nil {
		respondWithInternalServerError(w, r, "", "ReplicasCachedListGet", err)
		return
	}
	namespaceDeploymentReplica := NamespaceDeploymentReplica{Namespace: nsName, Deployment: dName, Replicas: deploymentList.Replicas}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(namespaceDeploymentReplica)
}

// Endpoint #3A
func serveReplicasAllGet(w http.ResponseWriter, r *http.Request) {
	self := "serveReplicasAllGet"
	klog.Infof("%s: entry", self)
	matches := reDeploymentAllReplicas.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		respondWithNotFound(w, r, "invalid arg(s)", r.URL.Path)
		return
	}
	nsName := matches[1]
	klog.Infof("%s: nsName=%q", self, nsName)
	if !NamespaceCachedExists(false, nsName) {
		respondWithNotFound(w, r, "namespace not found", nsName)
		return
	}
	deploymentList, err := ReplicasCachedListAllGet(false, nsName)
	if err != nil {
		respondWithInternalServerError(w, r, "", "ReplicasCachedListAllGet", err)
		return
	}
	namespaceListItem := NamespaceListItem{Name: nsName, Deployments: deploymentList}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(namespaceListItem)
}

// Endpoint #4
func serveReplicasSet(w http.ResponseWriter, r *http.Request) {
	self := "serveReplicasSet"
	klog.Infof("%s: entry", self)
	matches := reDeploymentSetReplicas.FindStringSubmatch(r.URL.Path)
	if len(matches) < 4 {
		respondWithNotFound(w, r, "invalid arg(s)", r.URL.Path)
		return
	}
	nsName, dName, replicas := matches[1], matches[2], matches[3]
	klog.Infof("%s: nsName=%q;  dName=%q;  replicas=%v", self, nsName, dName, replicas)
	if !NamespaceCachedExists(false, nsName) {
		respondWithNotFound(w, r, "namespace not found", nsName)
		return
	}
	if !DeploymentCachedExists(false, nsName, dName) {
		respondWithNotFound(w, r, "deployment not found", fmt.Sprintf("%s/%s", nsName, dName))
		return
	}
	replicasI, _ := strconv.Atoi(replicas)
	err := ReplicasSet(HttpSavedApp, nsName, dName, replicasI)
	if err != nil {
		respondWithInternalServerError(w, r, "", "replicasSet", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Endpoint #5
func serveLiveness(w http.ResponseWriter, r *http.Request) {
	self := "serveLiveness"
	klog.Infof("%s: entry", self)
	w.WriteHeader(http.StatusOK)
}

// Endpoint #6
func serveReadiness(w http.ResponseWriter, r *http.Request) {
	self := "serveReadiness"
	klog.Infof("%s: entry", self)
	w.WriteHeader(http.StatusOK)
}

func initHttp(App *AppX) error {
	self := "initHttp"
	klog.Infof("%s: entry", self)
	App.Mux = http.NewServeMux()
	h := &handler{}
	App.Mux.Handle("/", h)

	HttpSavedApp = App
	return nil
}
