#!/usr/bin/env bash
minikube delete
minikube start
make manifests/crd.yaml
make image-chaos-daemon
make image-chaos-mesh
make docker-push
./loadimages_minikube.sh
./installchaosmesh_helm.sh
