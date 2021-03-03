#!/usr/bin/env bash
make manifests/crd.yaml
make image-chaos-daemon
make image-chaos-mesh
make docker-push
minikube delete
minikube start
./loadimages_minikube.sh
minikube delete
minikube start
./loadimages_minikube.sh
./installchaosmesh_helm.sh
