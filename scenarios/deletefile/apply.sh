#!/usr/bin/env bash
cd "$(dirname "$0")"
minikube ssh 'sudo mkdir /mnt/data && sudo chown -R :1234 /mnt/data/ && sudo chmod g+rw /mnt/data'
kubectl apply -f deployment.yaml
