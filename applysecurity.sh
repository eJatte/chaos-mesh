#!/usr/bin/env bash
kubectl apply -f nonrootpsp.yaml
kubectl apply -f service_account_cr.yaml
kubectl apply -f service_account_crb.yaml
minikube stop
minikube start --extra-config=apiserver.enable-admission-plugins=PodSecurityPolicy
