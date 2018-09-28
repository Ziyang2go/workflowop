#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/Ziyang2go/workflowop/pkg/generated \
github.com/Ziyang2go/workflowop/pkg/apis \
threekit:v1alpha \
--go-header-file "./tmp/codegen/boilerplate.go.txt"
