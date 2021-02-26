#!/usr/bin/env bash
minikube delete
minikube start
./createandpushimage.sh
./loadimages_minikube.sh
./installchaosmesh_helm.sh
