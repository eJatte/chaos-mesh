#!/usr/bin/env bash
./deletekindcluster.sh
./startkindcluster.sh
./createandpushimage.sh
./loadimages_kind.sh
./installchaosmesh_helm.sh
