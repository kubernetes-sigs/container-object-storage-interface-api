#! /bin/bash
#
# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# This script runs inside a Prow job. It can run unit tests ("make test")
# and E2E testing. This E2E testing covers different scenarios. 
# 
# The intended usage of this script is that individual repos import
# release-tools, then link their top-level prow.sh to this or
# include it in that file. When including it, several of the variables
# can be overridden in the top-level prow.sh to customize the script
# for the repo.
#
# The expected environment is:
# - $GOPATH/src/<import path> for the repository that is to be tested,
#   with PR branch merged (when testing a PR)
# - running on linux-amd64
# - bazel installed (when testing against Kubernetes master), must be recent
#   enough for Kubernetes master
# - kind (https://github.com/kubernetes-sigs/kind) installed
# - optional: Go already installed

RELEASE_TOOLS_ROOT="$(realpath "$(dirname "${BASH_SOURCE[0]}")")"
REPO_DIR="$(pwd)"

# Sets the default value for a variable if not set already and logs the value.
# Any variable set this way is usually something that a repo's .prow.sh
# or the job can set.
configvar () {
    # Ignore: Word is of the form "A"B"C" (B indicated). Did you mean "ABC" or "A\"B\"C"?
    # shellcheck disable=SC2140
    eval : \$\{"$1":="\$2"\}
    eval echo "\$3:" "$1=\${$1}"
}

# Takes the minor version of $COSI_PROW_KUBERNETES_VERSION and overrides it to
# $1 if they are equal minor versions. Ignores versions that begin with
# "release-".
override_k8s_version () {
    local current_minor_version
    local override_minor_version

    # Ignore: See if you can use ${variable//search/replace} instead.
    # shellcheck disable=SC2001
    current_minor_version="$(echo "${COSI_PROW_KUBERNETES_VERSION}" | sed -e 's/\([0-9]*\)\.\([0-9]*\).*/\1\.\2/')"

    # Ignore: See if you can use ${variable//search/replace} instead.
    # shellcheck disable=SC2001
    override_minor_version="$(echo "${1}" | sed -e 's/\([0-9]*\)\.\([0-9]*\).*/\1\.\2/')"
    if [ "${current_minor_version}" == "${override_minor_version}" ]; then
      COSI_PROW_KUBERNETES_VERSION="$1"
      echo "Overriding COSI_PROW_KUBERNETES_VERSION with $1: $COSI_PROW_KUBERNETES_VERSION"
    fi
}

# Prints the value of a variable + version suffix, falling back to variable + "LATEST".
get_versioned_variable () {
    local var="$1"
    local version="$2"
    local value

    eval value="\${${var}_${version}}"
    if ! [ "$value" ]; then
        eval value="\${${var}_LATEST}"
    fi
    echo "$value"
}

# If we have a vendor directory, then use it. We must be careful to only
# use this for "make" invocations inside the project's repo itself because
# setting it globally can break other go usages (like "go get <some command>"
# which is disabled with GOFLAGS=-mod=vendor).
configvar GOFLAGS_VENDOR "$( [ -d vendor ] && echo '-mod=vendor' )" "Go flags for using the vendor directory"

# Go versions can be specified seperately for different tasks
# If the pre-installed Go is missing or a different
# version, the required version here will get installed
# from https://golang.org/dl/.
go_from_travis_yml () {
    grep "^ *- go:" "${RELEASE_TOOLS_ROOT}/travis.yml" | sed -e 's/.*go: *//'
}

configvar COSI_K8S_GO_VERSION "1.15.5" "This will override the k8s version, sometime k8s version is incorrect to fetch from go downloads"
configvar COSI_PROW_GO_VERSION_BUILD "$(go_from_travis_yml)" "Go version for building the component" # depends on component's source code
configvar COSI_PROW_GO_VERSION_E2E "${COSI_K8S_GO_VERSION}" "override Go version for building the Kubernetes E2E test suite" # normally doesn't need to be set, see install_e2e
configvar COSI_PROW_GO_VERSION_KIND "${COSI_PROW_GO_VERSION_BUILD}" "Go version for building 'kind'" # depends on COSI_PROW_KIND_VERSION below
configvar COSI_PROW_GO_VERSION_GINKGO "${COSI_PROW_GO_VERSION_BUILD}" "Go version for building ginkgo" # depends on COSI_PROW_GINKGO_VERSION below

# kind version to use. If the pre-installed version is different,
# the desired version is downloaded from https://github.com/kubernetes-sigs/kind/releases/download/
# (if available), otherwise it is built from source.
configvar COSI_PROW_KIND_VERSION "v0.6.0" "kind"

# ginkgo test runner version to use. If the pre-installed version is
# different, the desired version is built from source.
configvar COSI_PROW_GINKGO_VERSION v1.7.0 "Ginkgo"

# Ginkgo runs the E2E test in parallel. The default is based on the number
# of CPUs, but typically this can be set to something higher in the job.
configvar COSI_PROW_GINKO_PARALLEL "-p" "Ginko parallelism parameter(s)"

# Enables building the code in the repository. On by default, can be
# disabled in jobs which only use pre-built components.
configvar COSI_PROW_BUILD_JOB true "building code in repo enabled"

# Kubernetes version to test against. This must be a version number
# (like 1.13.3) for which there is a pre-built kind image (see
# https://hub.docker.com/r/kindest/node/tags), "latest" (builds
# Kubernetes from the master branch) or "release-x.yy" (builds
# Kubernetes from a release branch).
#
# This can also be a version that was not released yet at the time
# that the settings below were chose. The script will then
# use the same settings as for "latest" Kubernetes. This works
# as long as there are no breaking changes in Kubernetes, like
# deprecating or changing the implementation of an alpha feature.
configvar COSI_PROW_KUBERNETES_VERSION 1.24.0 "Kubernetes"

# This is a hack to workaround the issue that each version
# of kind currently only supports specific patch versions of
# Kubernetes. We need to override COSI_PROW_KUBERNETES_VERSION
# passed in by our CI/pull jobs to the versions that
# kind v0.5.0 supports.
#
# If the version is prefixed with "release-", then nothing
# is overridden.
override_k8s_version "1.24.0"

# COSI_PROW_KUBERNETES_VERSION reduced to first two version numbers and
# with underscore (1_13 instead of 1.13.3) and in uppercase (LATEST
# instead of latest).
#
# This is used to derive the right defaults for the variables below
# when a Prow job just defines the Kubernetes version.
cosi_prow_kubernetes_version_suffix="$(echo "${COSI_PROW_KUBERNETES_VERSION}" | tr . _ | tr '[:lower:]' '[:upper:]' | sed -e 's/^RELEASE-//' -e 's/\([0-9]*\)_\([0-9]*\).*/\1_\2/')"

# Work directory. It has to allow running executables, therefore /tmp
# is avoided. Cleaning up after the script is intentionally left to
# the caller.
configvar COSI_PROW_WORK "$(mkdir -p "$GOPATH/pkg" && mktemp -d "$GOPATH/pkg/cosiprow.XXXXXXXXXX")" "work directory"

# The E2E testing can come from an arbitrary repo. The expectation is that
# the repo supports "go test ./test/e2e -args --storage.testdriver" (https://github.com/kubernetes/kubernetes/pull/72836)
# after setting KUBECONFIG. As a special case, if the repository is Kubernetes,
# then `make WHAT=test/e2e/e2e.test` is called first to ensure that
# all generated files are present.
#
# COSI_PROW_E2E_REPO=none disables E2E testing.
# TOOO: remove versioned variables and make e2e version match k8s version
configvar COSI_PROW_E2E_VERSION_LATEST master "E2E version for Kubernetes master" # testing against Kubernetes master is already tracking a moving target, so we might as well use a moving E2E version
configvar COSI_PROW_E2E_REPO_LATEST https://github.com/kubernetes/kubernetes "E2E repo for Kubernetes >= 1.13.x" # currently the same for all versions
configvar COSI_PROW_E2E_IMPORT_PATH_LATEST k8s.io/kubernetes "E2E package for Kubernetes >= 1.13.x" # currently the same for all versions
configvar COSI_PROW_E2E_VERSION "$(get_versioned_variable COSI_PROW_E2E_VERSION "${cosi_prow_kubernetes_version_suffix}")"  "E2E version"
configvar COSI_PROW_E2E_REPO "$(get_versioned_variable COSI_PROW_E2E_REPO "${cosi_prow_kubernetes_version_suffix}")" "E2E repo"
configvar COSI_PROW_E2E_IMPORT_PATH "$(get_versioned_variable COSI_PROW_E2E_IMPORT_PATH "${cosi_prow_kubernetes_version_suffix}")" "E2E package"

# The version of dep to use for 'make test-vendor'. Ignored if the project doesn't
# use dep. Only binary releases of dep are supported (https://github.com/golang/dep/releases).
configvar COSI_PROW_DEP_VERSION v0.5.1 "golang dep version to be used for vendor checking"

// Version of the Spec used
configvar COSI_SPEC_VERSION master "version of the cosi spec will influence the crd object loaded for testing"

// Version of the API used
configvar COSI_API_VERSION master "version of the cosi api will influence the api objects loaded for testing"

// Version of the Controller used
configvar COSI_CONTROLLER_VERSION master "version of the cosi controller used for testing"

# TODO Each job can run one or more of the following tests, identified by
# a single word:
# - unit testing
# - parallel excluding alpha features
# - serial excluding alpha features
# - parallel, only alpha feature
# - serial, only alpha features
# - sanity
#
# Unknown or unsupported entries are ignored.
#
configvar COSI_PROW_TESTS "unit parallel serial parallel-alpha serial-alpha sanity" "tests to run"
tests_enabled () {
    local t1 t2
    # We want word-splitting here, so ignore: Quote to prevent word splitting, or split robustly with mapfile or read -a.
    # shellcheck disable=SC2206
    local tests=(${COSI_PROW_TESTS})
    for t1 in "$@"; do
        for t2 in "${tests[@]}"; do
            if [ "$t1" = "$t2" ]; then
                return
            fi
        done
    done
    return 1
}

tests_need_kind () {
    tests_enabled "parallel" "serial" "serial-alpha" "parallel-alpha" ||
        sanity_enabled
}
tests_need_non_alpha_cluster () {
    tests_enabled "parallel" "serial" ||
        sanity_enabled
}
tests_need_alpha_cluster () {
    tests_enabled "parallel-alpha" "serial-alpha"
}

# Serial vs. parallel is always determined by these regular expressions.
# Individual regular expressions are seperated by spaces for readability
# and expected to not contain spaces. Use dots instead. The complete
# regex for Ginkgo will be created by joining the individual terms.
configvar COSI_PROW_E2E_SERIAL '\[Serial\] \[Disruptive\]' "tags for serial E2E tests"
regex_join () {
    echo "$@" | sed -e 's/  */|/g' -e 's/^|*//' -e 's/|*$//' -e 's/^$/this-matches-nothing/g'
}

configvar COSI_PROW_E2E_SKIP 'Disruptive|different\s+node' "tests that need to be skipped"

configvar COSI_PROW_E2E_ALPHA "$(get_versioned_variable COSI_PROW_E2E_ALPHA "${cosi_prow_kubernetes_version_suffix}")" "alpha tests"

# This is the directory for additional result files. Usually set by Prow, but
# if not (for example, when invoking manually) it defaults to the work directory.
configvar ARTIFACTS "${COSI_PROW_WORK}/artifacts" "artifacts"
mkdir -p "${ARTIFACTS}"

run () {
    echo "$(date) $(go version | sed -e 's/.*version \(go[^ ]*\).*/\1/') $(if [ "$(pwd)" != "${REPO_DIR}" ]; then pwd; fi)\$" "$@" >&2
    "$@"
}

info () {
    echo >&2 INFO: "$@"
}

warn () {
    echo >&2 WARNING: "$@"
}

die () {
    echo >&2 ERROR: "$@"
    exit 1
}

# For additional tools.
COSI_PROW_BIN="${COSI_PROW_WORK}/bin"
mkdir -p "${COSI_PROW_BIN}"
PATH="${COSI_PROW_BIN}:$PATH"

# Ensure that PATH has the desired version of the Go tools, then run command given as argument.
# Empty parameter uses the already installed Go. In Prow, that version is kept up-to-date by
# bumping the container image regularly.
run_with_go () {
    local version
    version="$1"
    shift

    if ! [ "$version" ] || go version 2>/dev/null | grep -q "go$version"; then
        run "$@"
    else
        if ! [ -d "${COSI_PROW_WORK}/go-$version" ];  then
            run curl --fail --location "https://dl.google.com/go/go$version.linux-amd64.tar.gz" | tar -C "${COSI_PROW_WORK}" -zxf - || die "installation of Go $version failed"
            mv "${COSI_PROW_WORK}/go" "${COSI_PROW_WORK}/go-$version"
        fi
        PATH="${COSI_PROW_WORK}/go-$version/bin:$PATH" run "$@"
    fi
}

# Ensure that we have the desired version of kind.
install_kind () {
    if kind --version 2>/dev/null | grep -q " ${COSI_PROW_KIND_VERSION}$"; then
        return
    fi
    if run curl --fail --location -o "${COSI_PROW_WORK}/bin/kind" "https://github.com/kubernetes-sigs/kind/releases/download/${COSI_PROW_KIND_VERSION}/kind-linux-amd64"; then
        chmod u+x "${COSI_PROW_WORK}/bin/kind"
    else
        git_checkout https://github.com/kubernetes-sigs/kind "${GOPATH}/src/sigs.k8s.io/kind" "${COSI_PROW_KIND_VERSION}" --depth=1 &&
        (cd "${GOPATH}/src/sigs.k8s.io/kind" && make install INSTALL_DIR="${COSI_PROW_WORK}/bin")
    fi
}

# Ensure that we have the desired version of the ginkgo test runner.
install_ginkgo () {
    # COSI_PROW_GINKGO_VERSION contains the tag with v prefix, the command line output does not.
    if [ "v$(ginkgo version 2>/dev/null | sed -e 's/.* //')" = "${COSI_PROW_GINKGO_VERSION}" ]; then
        return
    fi
    git_checkout https://github.com/onsi/ginkgo "$GOPATH/src/github.com/onsi/ginkgo" "${COSI_PROW_GINKGO_VERSION}" --depth=1 &&
    # We have to get dependencies and hence can't call just "go build".
    run_with_go "${COSI_PROW_GO_VERSION_GINKGO}" go get github.com/onsi/ginkgo/ginkgo || die "building ginkgo failed" &&
    mv "$GOPATH/bin/ginkgo" "${COSI_PROW_BIN}"
}

# Ensure that we have the desired version of dep.
install_dep () {
    if dep version 2>/dev/null | grep -q "version:.*${COSI_PROW_DEP_VERSION}$"; then
        return
    fi
    run curl --fail --location -o "${COSI_PROW_WORK}/bin/dep" "https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64" &&
        chmod u+x "${COSI_PROW_WORK}/bin/dep"
}

# This checks out a repo ("https://github.com/kubernetes/kubernetes")
# in a certain location ("$GOPATH/src/k8s.io/kubernetes") at
# a certain revision (a hex commit hash, v1.13.1, master). It's okay
# for that directory to exist already.
git_checkout () {
    local repo path revision
    repo="$1"
    shift
    path="$1"
    shift
    revision="$1"
    shift

    mkdir -p "$path"
    if ! [ -d "$path/.git" ]; then
        run git init "$path"
    fi
    if (cd "$path" && run git fetch "$@" "$repo" "$revision"); then
        (cd "$path" && run git checkout FETCH_HEAD) || die "checking out $repo $revision failed"
    else
        # Might have been because fetching by revision is not
        # supported by GitHub (https://github.com/isaacs/github/issues/436).
        # Fall back to fetching everything.
        (cd "$path" && run git fetch "$repo" '+refs/heads/*:refs/remotes/cosiprow/heads/*' '+refs/tags/*:refs/tags/*') || die "fetching $repo failed"
        (cd "$path" && run git checkout "$revision") || die "checking out $repo $revision failed"
    fi
    # This is useful for local testing or when switching between different revisions in the same
    # repo.
    (cd "$path" && run git clean -fdx) || die "failed to clean $path"
}

# This clones a repo ("https://github.com/kubernetes/kubernetes")
# in a certain location ("$GOPATH/src/k8s.io/kubernetes") at
# a the head of a specific branch (i.e., release-1.13, master).
# The directory cannot exist.
git_clone_branch () {
    local repo path branch parent
    repo="$1"
    shift
    path="$1"
    shift
    branch="$1"
    shift

    parent="$(dirname "$path")"
    mkdir -p "$parent"
    (cd "$parent" && run git clone --single-branch --branch "$branch" "$repo" "$path") || die "cloning $repo" failed
    # This is useful for local testing or when switching between different revisions in the same
    # repo.
    (cd "$path" && run git clean -fdx) || die "failed to clean $path"
}

go_version_for_kubernetes () (
    local path="$1"
    local version="$2"
    local go_version

    # We use the minimal Go version specified for each K8S release (= minimum_go_version in hack/lib/golang.sh).
    # More recent versions might also work, but we don't want to count on that.
    go_version="$(grep minimum_go_version= "$path/hack/lib/golang.sh" | sed -e 's/.*=go//')"
    if ! [ "$go_version" ]; then
        die "Unable to determine Go version for Kubernetes $version from hack/lib/golang.sh."
    fi
    echo "$go_version"
)

cosi_prow_kind_have_kubernetes=false
# Brings up a Kubernetes cluster and sets KUBECONFIG.
# Accepts additional feature gates in the form gate1=true|false,gate2=...
start_cluster () {
    local image gates
    gates="$1"

    if kind get clusters | grep -q cosi-prow; then
        run kind delete cluster --name=cosi-prow || die "kind delete failed"
    fi

    echo "build k/k source"
    # Build from source?
    if [[ "${COSI_PROW_KUBERNETES_VERSION}" =~ ^release-|^latest$ ]]; then
        if ! ${cosi_prow_kind_have_kubernetes}; then
            local version="${COSI_PROW_KUBERNETES_VERSION}"
            if [ "$version" = "latest" ]; then
                version=master
            fi
            git_clone_branch https://github.com/kubernetes/kubernetes "${COSI_PROW_WORK}/src/kubernetes" "$version" || die "checking out Kubernetes $version failed"

            go_version="$(go_version_for_kubernetes "${COSI_PROW_WORK}/src/kubernetes" "$version")" || die "cannot proceed without knowing Go version for Kubernetes"
            run_with_go "$go_version" kind build node-image --type bazel --image cosiprow/node:latest --kube-root "${COSI_PROW_WORK}/src/kubernetes" || die "'kind build node-image' failed"
            cosi_prow_kind_have_kubernetes=true
        fi
        image="cosiprow/node:latest"
    else
        image="kindest/node:v${COSI_PROW_KUBERNETES_VERSION}"
    fi
    cat >"${COSI_PROW_WORK}/kind-config.yaml" <<EOF
kind: Cluster
apiVersion: kind.sigs.k8s.io/v1alpha3
nodes:
- role: control-plane
- role: worker
- role: worker
EOF

    # kubeadm has API dependencies between apiVersion and Kubernetes version
    # 1.15+ requires kubeadm.k8s.io/v1beta2
    # We only run alpha tests against master so we don't need to maintain
    # different patches for different Kubernetes releases.
    if [[ -n "$gates" ]]; then
        cat >>"${COSI_PROW_WORK}/kind-config.yaml" <<EOF
kubeadmConfigPatches:
- |
  apiVersion: kubeadm.k8s.io/v1beta2
  kind: ClusterConfiguration
  metadata:
    name: config
  apiServer:
    extraArgs:
      "feature-gates": "$gates"
  controllerManager:
    extraArgs:
      "feature-gates": "$gates"
  scheduler:
    extraArgs:
      "feature-gates": "$gates"
- |
  apiVersion: kubeadm.k8s.io/v1beta2
  kind: InitConfiguration
  metadata:
    name: config
  nodeRegistration:
    kubeletExtraArgs:
      "feature-gates": "$gates"
- |
  apiVersion: kubelet.config.k8s.io/v1beta1
  kind: KubeletConfiguration
  metadata:
    name: config
  featureGates:
$(list_gates "$gates")
- |
  apiVersion: kubeproxy.config.k8s.io/v1alpha1
  kind: KubeProxyConfiguration
  metadata:
    name: config
  featureGates:
$(list_gates "$gates")
EOF
    fi

    info "kind-config.yaml:"
    cat "${COSI_PROW_WORK}/kind-config.yaml"
    echo "Creating kind cluster"
    if ! run kind create cluster --name cosi-prow --config "${COSI_PROW_WORK}/kind-config.yaml" --wait 5m --image "$image"; then
        echo "create cluster failed, $?"
        warn "Cluster creation failed. Will try again with higher verbosity."
        info "Available Docker images:"
        docker image ls
        if ! run kind --loglevel debug create cluster --retain --name cosi-prow --config "${COSI_PROW_WORK}/kind-config.yaml" --wait 5m --image "$image"; then
            run kind export logs --name cosi-prow "$ARTIFACTS/kind-cluster"
            die "Cluster creation failed again, giving up. See the 'kind-cluster' artifact directory for additional logs."
        fi
    fi
    export KUBECONFIG="${HOME}/.kube/config"
}

# Deletes kind cluster inside a prow job
delete_cluster_inside_prow_job() {
    # Inside a real Prow job it is better to clean up at runtime
    # instead of leaving that to the Prow job cleanup code
    # because the later sometimes times out (https://github.com/kubernetes-COSI/COSI-release-tools/issues/24#issuecomment-554765872).
    if [ "$JOB_NAME" ]; then
        if kind get clusters | grep -q cosi-prow; then
            run kind delete cluster --name=cosi-prow || die "kind delete failed"
        fi
        unset KUBECONFIG
    fi
}

# Looks for the deployment as specified by COSI_PROW_DEPLOYMENT and COSI_PROW_KUBERNETES_VERSION
# in the given directory.
find_deployment () {
    local dir file
    dir="$1"

    # Fixed deployment name? Use it if it exists, otherwise fail.
    if [ "${COSI_PROW_DEPLOYMENT}" ]; then
        file="$dir/${COSI_PROW_DEPLOYMENT}/deploy.sh"
        if ! [ -e "$file" ]; then
            return 1
        fi
        echo "$file"
        return 0
    fi

    # Ignore: See if you can use ${variable//search/replace} instead.
    # shellcheck disable=SC2001
    file="$dir/kubernetes-$(echo "${COSI_PROW_KUBERNETES_VERSION}" | sed -e 's/\([0-9]*\)\.\([0-9]*\).*/\1.\2/')/deploy.sh"
    if ! [ -e "$file" ]; then
        file="$dir/kubernetes-latest/deploy.sh"
        if ! [ -e "$file" ]; then
            return 1
        fi
    fi
    echo "$file"
}

# This installs the cosi driver. It's called with a list of env variables
# that override the default images. COSI_PROW_DRIVER_CANARY overrides all
# image versions with that canary version.
install_cosi_driver () {
#    local images deploy_driver
    images="$*"
}

kubectl_apply () {
  // TODO once this CRD is part of core replace it with  'kubectl apply -f $1 --validate=false'
  curl $1 | sed '/annotations/ a  \   \ "api-approved.kubernetes.io": "https://github.com/kubernetes-sigs/container-object-storage-interface-api/pull/2"' | kubectl apply -f - --validate=false
}

# Installs all nessesary CRDs  
install_crds() {
  # Wait until cosi CRDs are in place.
  CRD_BASE_DIR="https://raw.githubusercontent.com/kubernetes-sigs/container-object-storage-interface-api/${COSI_SPEC_VERSION}/crds"
  kubectl_apply  "${CRD_BASE_DIR}/objectstorage.k8s.io_bucketclasses.yaml"
  kubectl_apply  "${CRD_BASE_DIR}/objectstorage.k8s.io_bucketrequests.yaml"
  kubectl_apply  "${CRD_BASE_DIR}/objectstorage.k8s.io_buckets.yaml"
  kubectl_apply  "${CRD_BASE_DIR}/objectstorage.k8s.io_bucketaccessclasses.yaml"
  kubectl_apply  "${CRD_BASE_DIR}/objectstorage.k8s.io_bucketaccessrequests.yaml"
  kubectl_apply  "${CRD_BASE_DIR}/objectstorage.k8s.io_bucketaccesses.yaml"
  cnt=0
  until kubectl get bucketaccessclasses.objectstorage.k8s.io \
    && kubectl get bucketaccessrequests.objectstorage.k8s.io \
    && kubectl get bucketaccesses.objectstorage.k8s.io \
    && kubectl get bucketclasses.objectstorage.k8s.io	 \
    && kubectl get bucketrequests.objectstorage.k8s.io  \
    && kubectl get buckets.objectstorage.k8s.io; do
    if [ $cnt -gt 30 ]; then
        echo >&2 "ERROR: cosi CRDs not ready after over 1 min"
        exit 1
    fi
    echo "$(date +%H:%M:%S)" "waiting for cosi CRDs, attempt #$cnt"
	cnt=$((cnt + 1))
    sleep 2
  done
}

# Install controller and associated RBAC, retrying until the pod is running.
install_controller() {
  kubectl apply -f "https://raw.githubusercontent.com/kubernetes-sigs/container-object-storage-interface-controller/${COSI_CONTROLLER_VERSION}/deploy/base/sa.yaml"
  kubectl apply -f "https://raw.githubusercontent.com/kubernetes-sigs/container-object-storage-interface-controller/${COSI_CONTROLLER_VERSION}/deploy/base/rbac.yaml"
  cnt=0
  until kubectl get clusterrolebinding objectstorage-controller; do
     if [ $cnt -gt 30 ]; then
        echo "Cluster role bindings:"
        kubectl describe clusterrolebinding
        echo >&2 "ERROR: controller RBAC not ready after over 5 min"
        exit 1
     fi
     echo "$(date +%H:%M:%S)" "waiting for cosi RBAC setup complete, attempt #$cnt"
	 cnt=$((cnt + 1))
     sleep 10   
  done


  kubectl apply -f "https://raw.githubusercontent.com/kubernetes-sigs/container-object-storage-interface-controller/${COSI_CONTROLLER_VERSION}/deploy/base/deployment.yaml"
  cnt=0
  kubectl get pods 
  kubectl get pods -l app=objectstorage-controller 
  expected_running_pods=$(curl "https://raw.githubusercontent.com/kubernetes-sigs/container-object-storage-interface-controller/${COSI_CONTROLLER_VERSION}/deploy/base/deployment.yaml" | grep replicas | cut -d ':' -f 2-)
  while [ "$(kubectl get pods -l app.kubernetes.io/name=container-object-storage-interface-controller | grep 'Running' -c)" -lt "$expected_running_pods" ]; do
    if [ $cnt -gt 30 ]; then
        echo "objectstorage-controller pod status:"
        kubectl describe pods -l app.kubernetes.io/name=container-object-storage-interface-controller
        echo >&2 "ERROR: cosi controller not ready after over 5 min"
        exit 1
    fi
    echo "$(date +%H:%M:%S)" "waiting for cosi controller deployment to complete, attempt #$cnt"
	cnt=$((cnt + 1))
    sleep 10   
  done
}

# collect logs and cluster status (like the version of all components, Kubernetes version, test version)
collect_cluster_info () {
    cat <<EOF
=========================================================
Kubernetes:
$(kubectl version)

Driver installation in default namespace:
$(kubectl get all)


=========================================================
EOF

}

# Gets logs of all containers in all namespaces. When passed -f, kubectl will
# keep running and capture new output. Prints the pid of all background processes.
# The caller must kill (when using -f) and/or wait for them.
#
# May be called multiple times and thus appends.
start_loggers () {
    kubectl get pods --all-namespaces -o go-template --template='{{range .items}}{{.metadata.namespace}} {{.metadata.name}} {{range .spec.containers}}{{.name}} {{end}}{{"\n"}}{{end}}' | while read -r namespace pod containers; do
        for container in $containers; do
            mkdir -p "${ARTIFACTS}/$namespace/$pod"
            kubectl logs -n "$namespace" "$@" "$pod" "$container" >>"${ARTIFACTS}/$namespace/$pod/$container.log" &
            echo "$!"
        done
    done
}

# Makes the E2E test suite binary available as "${COSI_PROW_WORK}/e2e.test".
install_e2e () {
    if [ -e "${COSI_PROW_WORK}/e2e.test" ]; then
        return
    fi

    git_checkout "${COSI_PROW_E2E_REPO}" "${GOPATH}/src/${COSI_PROW_E2E_IMPORT_PATH}" "${COSI_PROW_E2E_VERSION}" --depth=1 &&
    if [ "${COSI_PROW_E2E_IMPORT_PATH}" = "k8s.io/kubernetes" ]; then
        go_version="${COSI_PROW_GO_VERSION_E2E:-$(go_version_for_kubernetes "${GOPATH}/src/${COSI_PROW_E2E_IMPORT_PATH}" "${COSI_PROW_E2E_VERSION}")}" &&
        run_with_go "$go_version" make WHAT=test/e2e/e2e.test "-C${GOPATH}/src/${COSI_PROW_E2E_IMPORT_PATH}" &&
        ln -s "${GOPATH}/src/${COSI_PROW_E2E_IMPORT_PATH}/_output/bin/e2e.test" "${COSI_PROW_WORK}"
    else
        run_with_go "${COSI_PROW_GO_VERSION_E2E}" go test -c -o "${COSI_PROW_WORK}/e2e.test" "${COSI_PROW_E2E_IMPORT_PATH}/test/e2e"
    fi
}

# Captures pod output while running some other command.
run_with_loggers () (
    loggers=$(start_loggers -f)
    trap 'kill $loggers' EXIT

    run "$@"
)

# Invokes the filter-junit.go tool.
run_filter_junit () {
    run_with_go "${COSI_PROW_GO_VERSION_BUILD}" go run "${RELEASE_TOOLS_ROOT}/filter-junit.go" "$@"
}

# Runs the E2E test suite in a sub-shell.
run_e2e () (
    name="$1"
    shift

    install_e2e || die "building e2e.test failed"
    install_ginkgo || die "installing ginkgo failed"

    # Rename, merge and filter JUnit files. Necessary in case that we run the E2E suite again
    # and to avoid the large number of "skipped" tests that we get from using
    # the full Kubernetes E2E testsuite while only running a few tests.
    move_junit () {
        if ls "${ARTIFACTS}"/junit_[0-9]*.xml 2>/dev/null >/dev/null; then
            run_filter_junit -t="External Storage" -o "${ARTIFACTS}/junit_${name}.xml" "${ARTIFACTS}"/junit_[0-9]*.xml && rm -f "${ARTIFACTS}"/junit_[0-9]*.xml
        fi
    }
    trap move_junit EXIT

    cd "${GOPATH}/src/${COSI_PROW_E2E_IMPORT_PATH}" &&
    run_with_loggers ginkgo -v "$@" "${COSI_PROW_WORK}/e2e.test" -- -report-dir "${ARTIFACTS}"
# add this flag so that we cn switch to run tests against any driver -storage.testdriver="${COSI_PROW_WORK}/test-driver.yaml"
)

ascii_to_xml () {
    # We must escape special characters and remove escape sequences
    # (no good representation in the simple XML that we generate
    # here). filter_junit.go would choke on them during decoding, even
    # when disabling strict parsing.
    sed -e 's/&/&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' -e 's/\x1B...//g'
}

# The "make test" output starts each test with "### <test-target>:"
# and then ends when the next test starts or with "make: ***
# [<test-target>] Error 1" when there was a failure. Here we read each
# line of that output, split it up into individual tests and generate
# a make-test.xml file in JUnit format.
make_test_to_junit () {
    local ret out testname testoutput
    ret=0
    # Plain make-test.xml was not delivered as text/xml by the web
    # server and ignored by spyglass. It seems that the name has to
    # match junit*.xml.
    out="${ARTIFACTS}/junit_make_test.xml"
    testname=
    echo "<testsuite>" >>"$out"

    while IFS= read -r line; do
        echo "$line" # pass through
        if echo "$line" | grep -q "^### [^ ]*:$"; then
            if [ "$testname" ]; then
                # previous test succesful
                echo "    </system-out>" >>"$out"
                echo "  </testcase>" >>"$out"
            fi
            # Ignore: See if you can use ${variable//search/replace} instead.
            # shellcheck disable=SC2001
            #
            # start new test
            testname="$(echo "$line" | sed -e 's/^### \([^ ]*\):$/\1/')"
            testoutput=
            echo "  <testcase name=\"$testname\">" >>"$out"
            echo "    <system-out>" >>"$out"
        elif echo "$line" | grep -q '^make: .*Error [0-9]*$'; then
            if [ "$testname" ]; then
                # Ignore: Consider using { cmd1; cmd2; } >> file instead of individual redirects.
                # shellcheck disable=SC2129
                #
                # end test with failure
                echo "    </system-out>" >>"$out"
                # Include the same text as in <system-out> also in <failure>,
                # because then it is easier to view in spyglass (shown directly
                # instead of having to click through to stdout).
                echo "    <failure>" >>"$out"
                echo -n "$testoutput" | ascii_to_xml >>"$out"
                echo "    </failure>" >>"$out"
                echo "  </testcase>" >>"$out"
            fi
            # remember failure for exit code
            ret=1
            # not currently inside a test
            testname=
        else
            if [ "$testname" ]; then
                # Test output.
                echo "$line" | ascii_to_xml >>"$out"
                testoutput="$testoutput$line
"
            fi
        fi
    done
    # if still in a test, close it now
    if [ "$testname" ]; then
        echo "    </system-out>" >>"$out"
        echo "  </testcase>" >>"$out"
    fi
    echo "</testsuite>" >>"$out"

    # this makes the error more visible in spyglass
    if [ "$ret" -ne 0 ]; then
        echo "ERROR: 'make test' failed"
        return 1
    fi
}

# version_gt returns true if arg1 is greater than arg2.
#
# This function expects versions to be one of the following formats:
#   X.Y.Z, release-X.Y.Z, vX.Y.Z
#
#   where X,Y, and Z are any number.
#
# Partial versions (1.2, release-1.2) work as well.
# The follow substrings are stripped before version comparison:
#   - "v"
#   - "release-"
#   - "kubernetes-"
#
# Usage:
# version_gt release-1.3 v1.2.0  (returns true)
# version_gt v1.1.1 v1.2.0  (returns false)
# version_gt 1.1.1 v1.2.0  (returns false)
# version_gt 1.3.1 v1.2.0  (returns true)
# version_gt 1.1.1 release-1.2.0  (returns false)
# version_gt 1.2.0 1.2.2  (returns false)
function version_gt() { 
    versions=$(for ver in "$@"; do ver=${ver#release-}; ver=${ver#kubernetes-}; echo "${ver#v}"; done)
    greaterVersion=${1#"release-"};
    greaterVersion=${greaterVersion#"kubernetes-"};
    greaterVersion=${greaterVersion#"v"};
    test "$(printf '%s' "$versions" | sort -V | head -n 1)" != "$greaterVersion"
}

main () {
    local images ret
    ret=0

    images=
    if ${COSI_PROW_BUILD_JOB}; then
        # A successful build is required for testing.
        run_with_go "${COSI_PROW_GO_VERSION_BUILD}" make all "GOFLAGS_VENDOR=${GOFLAGS_VENDOR}" || die "'make all' failed"
        # We don't want test failures to prevent E2E testing below, because the failure
        # might have been minor or unavoidable, for example when experimenting with
        # changes in "release-tools" in a PR (that fails the "is release-tools unmodified"
        # test).
        if tests_enabled "unit"; then
            if [ -f Gopkg.toml ] && ! install_dep; then
                warn "installing 'dep' failed, cannot test vendoring"
                ret=1
            fi
            if ! run_with_go "${COSI_PROW_GO_VERSION_BUILD}" make -k test "GOFLAGS_VENDOR=${GOFLAGS_VENDOR}" 2>&1 | make_test_to_junit; then
                warn "'make test' failed, proceeding anyway"
                ret=1
            fi
        fi
        # Required for E2E testing.
        run_with_go "${COSI_PROW_GO_VERSION_BUILD}" make container "GOFLAGS_VENDOR=${GOFLAGS_VENDOR}" || die "'make container' failed"
    fi

    if tests_need_kind; then
        install_kind || die "installing kind failed"

        if ${COSI_PROW_BUILD_JOB}; then
            cmds="$(grep '^\s*CMDS\s*=' Makefile | sed -e 's/\s*CMDS\s*=//')"
            # Get the image that was just built (if any) from the
            # top-level Makefile CMDS variable and set the
            # deploy.sh env variables for it. We also need to
            # side-load those images into the cluster.
            for i in $cmds; do
                e=$(echo "$i" | tr '[:lower:]' '[:upper:]' | tr - _)
                images="$images ${e}_REGISTRY=quay.io/containerobjectstorage ${e}_TAG=cosiprow"

                # We must avoid the tag "latest" because that implies
                # always pulling the image
                # (https://github.com/kubernetes-sigs/kind/issues/328).
                docker tag "$i:latest" "$i:cosiprow" || die "tagging the locally built container image for $i failed"
            done

            if [ -e deploy/kubernetes/rbac.yaml ]; then
                # This is one of those components which has its own RBAC rules (like external-provisioner).
                # We are testing a locally built image and also want to test with the the current,
                # potentially modified RBAC rules.
                if [ "$(echo "$cmds" | wc -w)" != 1 ]; then
                    die "ambiguous deploy/kubernetes/rbac.yaml: need exactly one command, got: $cmds"
                fi
                e=$(echo "$cmds" | tr '[:lower:]' '[:upper:]' | tr - _)
                images="$images ${e}_RBAC=$(pwd)/deploy/kubernetes/rbac.yaml"
            fi
        fi

        if tests_need_non_alpha_cluster; then
            start_cluster || die "starting the non-alpha cluster failed"

            # Install necessary CRDs and controllers
            # For Kubernetes 1.19+, we will install the CRDs and controller.
            if version_gt "${COSI_PROW_KUBERNETES_VERSION}" "1.16.255" || "${COSI_PROW_KUBERNETES_VERSION}" == "latest"; then
                info "Version ${COSI_PROW_KUBERNETES_VERSION}, installing CRDs and cosi controller"
                install_crds
                install_controller
            else
                info "Version ${COSI_PROW_KUBERNETES_VERSION}, skipping CRDs and cosi controller"
            fi

            collect_cluster_info 

            if tests_enabled "parallel"; then
                if ! run_e2e parallel ${COSI_PROW_GINKO_PARALLEL} \
                     -focus="ObjectStorage" \
                     -skip="$(regex_join "${COSI_PROW_E2E_SERIAL}" "${COSI_PROW_E2E_ALPHA}" "${COSI_PROW_E2E_SKIP}")"; then
                    warn "E2E parallel failed"
                    ret=1
                fi
            fi

            if tests_enabled "serial"; then
                if ! run_e2e serial \
                     -focus="ObjectStorage.*" \
                     -skip="$(regex_join "${COSI_PROW_E2E_ALPHA}" "${COSI_PROW_E2E_SKIP}")"; then
                    warn "E2E serial failed"
                    ret=1
                fi
            fi
 
        fi
        delete_cluster_inside_prow_job
    fi
    # Merge all junit files into one. This gets rid of duplicated "skipped" tests.
    if ls "${ARTIFACTS}"/junit_*.xml 2>/dev/null >&2; then
        run_filter_junit -o "${COSI_PROW_WORK}/junit_final.xml" "${ARTIFACTS}"/junit_*.xml && rm "${ARTIFACTS}"/junit_*.xml && mv "${COSI_PROW_WORK}/junit_final.xml" "${ARTIFACTS}"
    fi

    return "$ret"
}
