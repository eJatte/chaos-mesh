#!/usr/bin/env bash
minikube cache add localhost:5000/pingcap/chaos-mesh:latest
minikube cache add localhost:5000/pingcap/chaos-daemon:latest
minikube cache add localhost:5000/pingcap/chaos-dashboard:latest
