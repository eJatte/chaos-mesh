#!/usr/bin/env bash
kubectl create namespace chaos-testing
kubectl apply -f manifests/
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock --set dashboard.securityMode=false
