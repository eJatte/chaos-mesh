#!/usr/bin/env bash
kubectl delete namespace chaos-testing
kubectl create namespace chaos-testing
kubectl apply -f manifests/
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set dashboard.create=true --set dashboard.securityMode=false
