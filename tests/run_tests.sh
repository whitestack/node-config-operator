#!/bin/bash

# Check if the first parameter is set, otherwise default to 3
NUM_NODES="${1:-3}"

# Check if the second argument is set, otherwise default to "config_files/nodeconfig-sample.yaml"
CONFIG_FILE="${2:-"config_files/nodeconfig-sample.yaml"}"

# Define directories, and file paths
OPERATOR_DIR="../"
OPERATOR_CHART_DIR="../chart"
CONFIGFILES_DIR="config_files/"
KUBECONFIG=".kube/minikube-config"

export KUBECONFIG

# Check if the specified test script file exists
if [ -e "$CONFIG_FILE" ]; then
    NODECONFIG_FILE="$CONFIG_FILE"
elif [ -e "$CONFIGFILES_DIR$CONFIG_FILE" ]; then
    NODECONFIG_FILE="$CONFIGFILES_DIR$CONFIG_FILE"
elif [ -e "$OPERATOR_DIR$CONFIG_FILE" ]; then
    NODECONFIG_FILE="$OPERATOR_DIR$CONFIG_FILE"
else
    echo "❌  Error: The specified node config file does not exist: $CONFIG_FILE"
    exit 1
fi

# Function to check if a command exists
command_exists() {
    command -v "$@" > /dev/null 2>&1
}

# Function to start Minikube
start_minikube() {
    local num_nodes=$1
    echo "📦  Starting Minikube cluster with $num_nodes nodes..."
    if ! minikube start --nodes=$num_nodes --apiserver-port=6443; then
        echo "❌  Error: Failed to start Minikube cluster."
        exit 1
    fi

    # Load image to cluster
	  minikube image load node-config-operator:test
}

# Function to verify cluster accessibility and functionality
verify_cluster() {
    echo "➖  Verifying cluster accessibility and functionality..."
    # Display the status while waiting
    while true; do
        if minikube status &> /dev/null; then
            echo "✅  Cluster is accessible and functional."
            break
        else
            echo "➖🕐 Cluster is not yet accessible or not functional. Waiting..."
            sleep 10
        fi
    done
}

# Function to create namespace and context
create_ns_context(){
    echo "🖌️  Creating namespace nco-tests-ns..."
    kubectl create namespace nco-tests-ns
    echo "🖌️  Creating context nco-tests-context..."
    kubectl config set-context nco-tests-context --cluster=minikube --user=minikube
    kubectl config use-context nco-tests-context &> /dev/null
}

# Function to deploy the node-config-operator
deploy_operator() {
    echo "🔨  Deploying node-config-operator using Helm..."
    helm install node-config-operator "${OPERATOR_CHART_DIR}" \
        -n nco-tests-ns \
        --kube-context=nco-tests-context \
        --set managerConfig.hostfsEnabled=true \
        --set controllerManager.manager.image.tag=test \
        --set controllerManager.manager.image.repository=node-config-operator
}

# Function to run tests
run_tests() {
    echo "🔩  Applying test node configurations and waiting before executing tests..."
    kubectl apply -f "$NODECONFIG_FILE" -n nco-tests-ns
    sleep 60
    echo "📄  Running specific tests..."
    chmod +x tests.sh
    ./tests.sh $NODECONFIG_FILE
}

# Function to perform cleanup
cleanup() {
    echo "🧹  Cleaning up Minikube cluster..."
    minikube delete
    rm -R ".kube"
}

# Trap to cleanup on EXIT
#trap cleanup EXIT

# Main execution block

echo "🚦  Starting the test setup..."

# Install kubectl if not installed
if ! command_exists kubectl; then
    echo "❓  kubectl is not installed. Installing..."
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
fi

# Function to install Minikube if not installed
if ! command_exists minikube; then
    echo "❓  Minikube is not installed. Installing..."
    curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
    && sudo install minikube-linux-amd64 /usr/local/bin/minikube
    sudo chown $(whoami) /usr/local/bin/minikube
fi

# Function to install helm if not installed
if ! command_exists helm; then
    curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
    chmod 700 get_helm.sh
    ./get_helm.sh
fi

# Function to install yq if not installed
if ! command_exists yq; then
    echo "❓  yq is not installed. Installing..."
    sudo wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/bin/yq &&\
    sudo chmod +x /usr/bin/yq
fi

# Start Minikube
start_minikube $NUM_NODES

# Verify the cluster
verify_cluster

# Create namespace and context
create_ns_context

# Deploy the node-config-operator
deploy_operator

# Run the tests
run_tests

echo "🔚  Tests completed."
