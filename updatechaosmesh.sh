#!/usr/bin/env bash
./createandpushimage.sh
./loadimages_kind.sh
helm upgrade chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set dashboard.create=true --set dashboard.securityMode=false
