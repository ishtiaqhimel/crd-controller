#!/bin/bash

set -x

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

vendor/k8s.io/code-generator/generate-groups.sh all \
  github.com/ishtiaqhimel/crd-controller/pkg/client \
  github.com/ishtiaqhimel/crd-controller/pkg/apis \
  crd.com:v1 \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt