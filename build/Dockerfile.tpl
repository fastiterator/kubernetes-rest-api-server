FROM ubuntu:23.10
LABEL maintainer="Mark Epstein"
USER root
COPY ${DIR_COMMAND}/${COMMAND}/${COMMAND}-${ARCHDESC} /usr/local/bin/${COMMAND}
RUN  chmod 755 /usr/local/bin/${COMMAND}
RUN  mkdir -p /users/nobody/.kube
COPY ${DIR_CONFIG}/kubectl-config.txt /users/nobody/.kube/config
RUN  mkdir -p /users/nobody/.minikube/profiles/minikube
COPY ${DIR_ASSET}/minikube/ca.crt /users/nobody/.minikube
COPY ${DIR_ASSET}/minikube/client.crt /users/nobody/.minikube/profiles/minikube
COPY ${DIR_ASSET}/minikube/client.key /users/nobody/.minikube/profiles/minikube
RUN  chmod 444 /users/nobody/.minikube/ca.crt
RUN  chmod 444 /users/nobody/.minikube/profiles/minikube/client.*
USER root
ENV HOME=/users/nobody
ENTRYPOINT ["/usr/local/bin/${COMMAND}"]
