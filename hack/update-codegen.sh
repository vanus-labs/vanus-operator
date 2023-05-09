set -o errexit
set -o nounset
set -o pipefail
set -x

# mkdir -p vendor/k8s.io
# git clone -b v0.27.0 git@github.com:kubernetes/code-generator.git vendor/k8s.io/code-generator

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
# TODO(jiangkai): automatically
CODEGEN_PKG=/Users/jiangkai/Projects/github/code-generator

${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/vanus-labs/vanus-operator/pkg/runtime/client github.com/vanus-labs/vanus-operator/pkg/runtime/apis \
  vanus:v1alpha1 \
  --output-base "${SCRIPT_ROOT}" \
  --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt"