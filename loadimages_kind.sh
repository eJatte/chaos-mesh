#!/usr/bin/env bash
kind load docker-image localhost:5000/pingcap/chaos-mesh:latest
kind load docker-image localhost:5000/pingcap/chaos-daemon:latest
kind load docker-image localhost:5000/pingcap/chaos-dashboard:latest
