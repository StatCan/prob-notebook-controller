# NOTE: You will need kubectl, helm, and kind installed.

PROJECT_NAME := istio-test
ISTIO_RELEASE := 1.5.10
REGISTRY := k8scc01covidacr.azurecr.io
CONTROLLER := prob-notebook-controller
TAG := latest

CHROMIUM := chromium
#CHROMIUM := flatpak run org.chromium.Chromium

# ProB, Unclass, unspecified
rstudio-alice_classification = --set classification="protected-b"
rstudio-bob_classification   = --set classification="unclassified"
rstudio-carla_classification =

# ProB, Unclass, unspecified
jupyter-alice_classification = --set classification="protected-b"
jupyter-bob_classification   = --set classification="unclassified"
jupyter-carla_classification =

.PHONY: kind create jupyter rstudio gatekeeper
.DEFAULT: create

create: kind create-users install-istio notebook-crd gatekeeper configure-users
	@echo "Cluster is all set up for users!"

kind:
	###  _   ___           _
	### | | / (_)         | |
	### | |/ / _ _ __   __| |
	### |    \| | '_ \ / _` |
	### | |\  \ | | | | (_| |
	### \_| \_/_|_| |_|\__,_|
	###
	mkdir -p .kube
	kind create cluster --name $(PROJECT_NAME)
	kubectl cluster-info --context kind-istio-test

clean delete:
	kind delete clusters $(PROJECT_NAME)

istio istio-$(ISTIO_RELEASE):
	wget -qO- "https://github.com/istio/istio/releases/download/$(ISTIO_RELEASE)/istio-$(ISTIO_RELEASE)-linux.tar.gz" | tar -xvz
	ln -s istio-$(ISTIO_RELEASE) istio

install-istio: istio
	### _____    _   _
	### |_   _|  | | (_)
	###   | | ___| |_ _  ___
	###   | |/ __| __| |/ _ \
	###  _| |___ \ |_| | (_) |
	###  \___/___/\__|_|\___/
	###
	# Install Istio
	kubectl create namespace istio-system || true
	./istio/bin/istioctl manifest apply \
		--set addonComponents.kiali.enabled=true \
		--set addonComponents.tracing.enabled=true || true

	kubectl apply -f istio-misc/gateway.yaml
	kubectl apply -f istio-misc/kiali-service.yaml

	kubectl create secret generic kiali -n istio-system \
		--from-literal "username=admin" --from-literal "passphrase=admin"

	# Replace LoadBalancer with ClusterIP
	sleep 10
	kubectl patch -n istio-system svc istio-ingressgateway \
		--type='json' -p='[{"op": "replace", "path": "/spec/type", "value":"NodePort"}]'

gatekeeper:
	kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/release-3.4/deploy/gatekeeper.yaml
	sleep 10
	while ! kubectl apply -f gatekeeper/gatekeeperconfig.yaml; do sleep 15; done
	kubectl apply -f gatekeeper/constraintTemplate.yaml
	while ! kubectl apply -f gatekeeper/constraint.yaml; do sleep 15; done

notebook-crd:
	kubectl apply -f https://raw.githubusercontent.com/kubeflow/kubeflow/master/components/notebook-controller/config/crd/bases/kubeflow.org_notebooks.yaml

chromium:
	$(CHROMIUM) --host-resolver-rules='MAP *.muchk8s.wow 127.0.0.1:8443' &
	kubectl port-forward -n istio-system svc/istio-ingressgateway 8443:80


controller:
	cd .. && docker build . -t $(REGISTRY)/$(CONTROLLER):$(TAG)
	kind load docker-image $(REGISTRY)/$(CONTROLLER):$(TAG) --name $(PROJECT_NAME)

controller-deploy: controller
	kubectl create namespace daaas || true
	kubectl label  namespace daaas istio-injection=enabled --overwrite || true
	IMAGE_SHA=$(TAG) envsubst < ../deploy/deploy.yaml.tpl > ../deploy/deploy.yaml
	kubectl apply -f ../deploy/deploy.yaml

###  _   _
### | | | |
### | | | |___  ___ _ __ ___
### | | | / __|/ _ \ '__/ __|
### | |_| \__ \  __/ |  \__ \
###  \___/|___/\___|_|  |___/
###
create-users: alice bob carla
alice bob carla:
	kubectl create namespace $@ || true
	kubectl label  namespace $@ istio-injection=enabled --overwrite || true

configure-users:
	# Alice has an Authorization Policy (Only Alice!)
	# Create the alice jupyter notebook
	kubectl apply -f alice-notebook-jupyter.yaml
	#This can now be managed by the controller
	# kubectl apply -f uploadPolicy.yaml
	# kubectl apply -f uploadPolicyJupyter.yaml

delete-users remove-users:
	kubectl delete namespace alice || true
	kubectl delete namespace bob   || true
	kubectl delete namespace carla || true

rstudio: rstudio-alice rstudio-bob rstudio-carla
rstudio-alice rstudio-bob rstudio-carla: create-users
	KUBEFLOW_USER=$$(echo $@ | awk -F '-' '{print $$NF}'); \
	helm template ./rstudio --set namespace=$$KUBEFLOW_USER \
		${${@}_classification} \
		--set host=kubeflow.muchk8s.wow | kubectl apply -f -

rstudio-delete: rstudio-delete-alice rstudio-delete-bob rstudio-delete-carla
rstudio-delete-alice rstudio-delete-bob rstudio-delete-carla:
	NS=$$(echo $@ | awk -F '-' '{print $$NF}'); \
	helm template ./rstudio --set namespace=$$NS | kubectl delete -f -

jupyter: jupyter-alice jupyter-bob jupyter-carla
jupyter-alice jupyter-bob jupyter-carla: create-users
	KUBEFLOW_USER=$$(echo $@ | awk -F '-' '{print $$NF}'); \
	helm template ./jupyter --set namespace=$$KUBEFLOW_USER \
		${${@}_classification} \
		--set host=kubeflow.muchk8s.wow | kubectl apply -f -

jupyter-delete: jupyter-delete-alice jupyter-delete-bob jupyter-delete-carla
jupyter-delete-alice jupyter-delete-bob jupyter-delete-carla:
	NS=$$(echo $@ | awk -F '-' '{print $$NF}'); \
	helm template ./jupyter --set namespace=$$NS | kubectl delete -f -
