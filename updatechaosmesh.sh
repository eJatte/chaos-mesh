#!/usr/bin/env bash
./createandpushimage.sh
./loadimages_kind.sh
helm upgrade chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock --set dashboard.securityMode=false
