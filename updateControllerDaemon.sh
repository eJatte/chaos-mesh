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
kubectl apply -f service_account_cr.yaml
kubectl apply -f service_account_crb.yaml
