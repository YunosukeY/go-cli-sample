#!/usr/bin/env bash

set -eu

usage() {
  cat <<USAGE
  Usage:
  - kind.sh create
  - kind.sh delete
USAGE
}

if [ "$#" != 1 ]; then
  usage
  exit 1
fi

command="$1"
repo_dir="$(git rev-parse --show-toplevel)"

if [ "$command" == "create" ]; then
  kind create cluster

  # deploy resources
  kubectl apply -f "${repo_dir}/k8s/namespace.yaml"
  kubectl apply -f "${repo_dir}/k8s/deployment.yaml"
  kubectl wait --for=condition=available deployment/mysql deployment/redis -n test-ns --timeout=10m
elif [ "$command" == "delete" ]; then
  kind delete cluster
else
  usage
  exit 1
fi
