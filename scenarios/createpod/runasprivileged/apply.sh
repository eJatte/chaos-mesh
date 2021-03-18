cd "$(dirname "$0")"
kubectl apply -f psp.yaml
minikube stop
minikube start --extra-config=apiserver.enable-admission-plugins=PodSecurityPolicy
