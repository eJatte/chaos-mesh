#!/usr/bin/env bash
cd "$(dirname "$0")"
config=$(cat config.yaml)
echo "$config"
minikube ssh 'sudo rm /var/lib/kubelet/config.yaml && sudo echo "'"$config"'" > config.yaml && sudo mv config.yaml /var/lib/kubelet/ && sudo systemctl daemon-reload && sudo systemctl restart kubelet.service'
