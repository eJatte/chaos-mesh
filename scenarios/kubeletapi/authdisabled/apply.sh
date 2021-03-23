#!/usr/bin/env bash
cd "$(dirname "$0")"
config=$(cat config.yaml)
node=$1
if [ $# -eq 0 ]
  then
    node="minikube"
fi
echo "$node"
echo "$config"
minikube ssh -n "$1" 'sudo rm /var/lib/kubelet/config.yaml && sudo echo "'"$config"'" > config.yaml && sudo mv config.yaml /var/lib/kubelet/ && sudo systemctl daemon-reload && sudo systemctl restart kubelet.service'
