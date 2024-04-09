help:
	@echo 'This is the Makefile for the Kubernetes REST API Server'
	@echo ''
	@echo ' DESCRIPTION'
	@echo '     Able to build and/or run:'
	@echo '         - server'
	@echo '         - test client'
	@echo '         - docker images'
	@echo '         - helm charts'
	@echo '         - minikube installation'
	@echo '         - real kubernetes installation'
	@echo '         - tests'
	@echo ''
	@echo ' USAGE'
	@echo '     1. Set env var: DH_USER, and optionally: COMMAND, GOOS, GOARCH, EKS_CLUSTER'
	@echo '            DH_USER       <your docker hub username>'
	@echo '            COMMAND       one of  ( server client ). Default: server'
	@echo '            GOOS          one of  ( linux darwin ). Default: linux'
	@echo '            GOARCH        one of  ( 386 mips ). Default: 386'
	@echo '            AWS_REGION    <name of AWS region to use>.  Default: us-west-2'
	@echo '            EKS_CLUSTER   <name of EKS cluster to create and/or use>.  Default: challenge-me'
	@echo '          Format:  "export DH_USER=<user> COMMAND=<command> ..."'
	@echo '          Example: "export DH_USER=justsomestuff COMMAND=server"'
	@echo ''
	@echo '     2. Run one or more make commands'
	@echo '          Format:  "make <command>"'
	@echo '          Example: "make push_docker"'
	@echo ''
	@echo ' PRIMARY TARGETS'
	@echo '     help              Show this help text'
	@echo '     clean             Clean up built objects'
	@echo '     build             Build go apps and Docker image'
	@echo '     push_docker       Push Docker image to Docker Hub'
	@echo '     run_docker_local  Run the Docker image locally, not in a Kubernetes cluster'
	@echo '     create_cluster    Create EKS cluster'

ifndef DH_USER
    $(error ERROR: env var "DH_USER" is not defined)
endif
ifeq ($(DH_USER),)
    $(error ERROR: env var "DH_USER" is empty)
endif

ifndef COMMAND
    COMMAND=server
endif
ifneq ($(COMMAND),server)
    ifneq ($(COMMAND),client)
        $(error ERROR: env var "COMMAND" not in  ( "client" "server" ): "$(COMMAND)")
    endif
endif

ifndef GOOS
    GOOS=linux
endif
ifneq ($(GOOS),darwin)
    ifneq ($(GOOS),linux)
        $(error ERROR: env var "GOOS" not in  ( "darwin" "linux" ): "$(GOOS)")
    endif
endif

ifndef GOARCH
    GOARCH=386
endif
ifneq ($(GOARCH),386)
    ifneq ($(GOARCH),mips)
        $(error ERROR: env var "GOARCH" not in  ( "386" "mips" ): "$(GOARCH)")
    endif
endif

ifndef TMPDIR
    TMPDIR=/tmp
endif

ifndef AWS_REGION
    AWS_REGION=us-west-2
endif

ifndef EKS_CLUSTER
    EKS_CLUSTER=challenge-me
endif


GENERATE_ID=$(shell od -vAn -N2 -tu2 < /dev/urandom)
SET_ID=$(eval ID=$(GENERATE_ID))
SELF=probable-potato-build-Makefile

DIR_ASSET=../assets
DIR_CONFIG=../configs
DIR_COMMAND=../cmd
DIR_BUILD=../build

ARCHBUILD=GOOS=$(GOOS) GOARCH=$(GOARCH)
ARCHDESC=$(GOOS)_$(GOARCH)

VARNAMES:=DIR_ASSET DIR_BUILD DIR_COMMAND DIR_CONFIG ARCHBUILD ARCHDESC COMMAND
SEDSTR:=$(foreach varname,$(VARNAMES),-e 's@\$\$${$(varname)}@$($(varname))@g')


.PHONY: build_go build_docker build fmt $(COMMAND)_clean build_clean clean \
	image_docker push_docker id_set compress_docker run_docker_local create_cluster

id_set:
	@$(SET_ID)

fmt: $(DIR_COMMAND)/$(COMMAND)/*.go
	@cd $(DIR_COMMAND)/$(COMMAND)  && \
	echo "** Formatting go program: $(COMMAND)"  && \
	go fmt  && \
	echo "**     [completed] - Formatted files matching: $(DIR_COMMAND)/$(COMMAND)/*.go"

build_go: $(DIR_COMMAND)/$(COMMAND)/*.go
	@cd $(DIR_COMMAND)/$(COMMAND)  && \
	echo "** Compiling go program: $(COMMAND)"  && \
	$(ARCHBUILD) go build -o $(COMMAND)-$(ARCHDESC)  && \
	echo "**     [completed] - Location: $(DIR_COMMAND)/$(COMMAND)-$(ARCHDESC)"

build_docker: build_go $(DIR_BUILD)/Dockerfile-$(COMMAND)-$(ARCHDESC)
	@cd $(DIR_BUILD)  && \
	echo "** Building Docker image for: $(COMMAND)"  && \
	docker build -q -t $(COMMAND)-$(ARCHDESC) -f $(DIR_BUILD)/Dockerfile-$(COMMAND)-$(ARCHDESC) .. > /dev/null && \
	echo "**     [completed] - Saved to local Docker image storage"
build: build_go build_docker

$(COMMAND)_clean:
	@cd $(DIR_COMMAND)/$(COMMAND)  && \
	echo "** Cleaning up: $(COMMAND)"  && \
	rm -f $(COMMAND)-* *~  && \
	echo "**     [completed] - Cleaned: $(COMMAND) in $(DIR_COMMAND)/$(COMMAND)"
build_clean:
	@cd $(DIR_BUILD)  && \
	echo "** Cleaning up build area: $(DIR_BUILD)"  && \
	rm -f *~ Dockerfile-*  && \
	echo "**     [completed] - Cleaned: $(DIR_BUILD)"
clean: $(COMMAND)_clean build_clean

$(DIR_BUILD)/Dockerfile-$(COMMAND)-$(ARCHDESC): $(DIR_BUILD)/Dockerfile.tpl
	@cd $(DIR_BUILD)  && \
	echo "** Creating Dockerfile for: $(COMMAND)"  && \
	sed $(SEDSTR) < $(DIR_BUILD)/Dockerfile.tpl > $(DIR_BUILD)/Dockerfile-$(COMMAND)-$(ARCHDESC)  && \
	echo "**     [completed] - Location: $(DIR_BUILD)/Dockerfile-$(COMMAND)-$(ARCHDESC)"

compress_docker: image_docker
	@cd $(DIR_BUILD)  && \
	echo "** Compressing Docker image for: $(COMMAND)"  && \
	gzip -c < $(DIR_BUILD)/Docker-image-$(COMMAND)-$(ARCHDESC) > $(DIR_BUILD)/Docker-image-$(COMMAND)-$(ARCHDESC).gz  && \
	echo "**     [completed] - Location: $(DIR_BUILD)/Docker-image-$(COMMAND)-$(ARCHDESC).gz"
$(DIR_BUILD)/Docker-image-$(COMMAND)-$(ARCHDESC).gz: compress_docker

image_docker: id_set build_docker
	@cd $(DIR_BUILD)  && \
	echo "** Tagging and Exporting Docker image for: $(COMMAND)"  && \
	docker images $(COMMAND)-$(ARCHDESC) | tail +2 | awk '{printf("%s",$$3)}' > $(TMPDIR)/$(SELF)-$(ID)  && \
	docker save -o Docker-image-$(COMMAND)-$(ARCHDESC) `cat $(TMPDIR)/$(SELF)-$(ID)`  && \
	docker tag `cat $(TMPDIR)/$(SELF)-$(ID)` $(DH_USER)/$(COMMAND)-$(ARCHDESC):latest  && \
	echo "**     [completed] - Tag: $(DH_USER)/$(COMMAND)-$(ARCHDESC):latest;  Image Location: $(DIR_BUILD)/Docker-image-$(COMMAND)-$(ARCHDESC)"  && \
	rm $(TMPDIR)/$(SELF)-$(ID)
Docker-image-$(COMMAND)-$(ARCHDESC): image_docker

push_docker: image_docker
	@echo "** Pushing image to Docker Hub"  && \
	docker push -q $(DH_USER)/$(COMMAND)-$(ARCHDESC):latest > /dev/null &&  \
	echo "**     [completed] Image name: docker.io/$(DH_USER)/$(COMMAND)-$(ARCHDESC):latest"

run_docker_local: push_docker
	@echo "** Running docker image: docker.io/$(DH_USER)/$(COMMAND)-$(ARCHDESC):latest"  && \
	docker run -i -t --network=host docker.io/$(DH_USER)/$(COMMAND)-$(ARCHDESC):latest

create_cluster:
	@echo "** Creating EKS cluster: $(EKS_CLUSTER) in region $(AWS_REGION)"  && \
	eksctl create cluster --name $(EKS_CLUSTER) --region $(AWS_REGION)  && \
	echo "**     [completed]: Region: $(AWS_REGION);  Cluster: $(EKS_CLUSTER)"

# other possible targets: dep, lint, vet, test, release, install
