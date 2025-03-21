#!/bin/bash

# Source the utility functions
source utils.sh

# The path to the node configuration file
NODECONFIG_FILE="$1"

# Path to the kubeconfig file
KUBECONFIG=".kube/minikube-config"

# Main execution block
echo "🔄  Retrieving all nodes in the cluster..."
# Get a list of all nodes in the cluster
NODES=$(kubectl get nodes -o name)

# Loop through each node in the cluster
for node in $NODES; do
    # Extract the node name from the full node path
    node_name=$(echo "$node" | cut -d '/' -f 2) # Strip the 'node/' prefix
    echo "➖  Checking node existence for $node_name..."
    # Check if the node exists
    kubectl get node "$node_name" &> /dev/null
    if [ $? -ne 0 ]; then
        echo "❔  Error: Node $node_name not found."
        exit 1
    fi
    echo "✅  Node $node_name found."

    # Extract specifications from the node configuration file
    specs=($(extract_spec_values $NODECONFIG_FILE))

    # Run test functions for each specification type
    for spec_type in "${specs[@]}"; do
        # Generate the name of the test function based on the specification type
        test_function="check_$(echo "$spec_type" | sed 's/\([A-Z]\)/_\1/g' | tr '[:upper:]' '[:lower:]' | sed 's/^_//')"
        # Execute the test function for the current node
        $test_function "$node_name"
    done
done

echo "🎉  All tests completed successfully."

# TODO: Add tests for status check. Currently the resource enters errors status
# because it can't run in minikube nodes. Wait until the tests are migrated to
# kubefire
