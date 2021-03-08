#!/usr/bin/env bash
cd "$(dirname "$0")"
minikube ssh 'sudo mkdir /mnt/data && cd /mnt/data/ && sudo sh -c "echo 'hello boys' > hello.txt"'
kubectl apply -f deployment.yaml
