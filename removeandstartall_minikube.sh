#!/usr/bin/env bash
minikube delete
minikube start
kubectl apply -f service_account_cr.yaml
kubectl apply -f service_account_crb.yaml
./createandpushimage.sh
./loadimages_minikube.sh
./installchaosmesh_helm.sh
