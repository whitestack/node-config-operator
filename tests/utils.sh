#!/bin/bash

# Function to run a command on a node
run_on_node() {
    local node_name=$1
    local cmd=$2

    # Run the command directly on the node using minikube ssh
    minikube ssh -n $node_name -- $cmd 2>&1
}

# Function to extract the 'spec' section from a YAML file
extract_spec_values() {
    local file="$1"
    local spec_section=$(yq eval '.spec' $file)
    # Parsing the spec section into individual values
    local specs=($(awk '/^[^[:space:]]/{sub(/:$/, ""); print $0}' <<< "$spec_section"))
    echo "${specs[@]}"
}

# Function to check kernel parameters on each node
check_kernel_parameters() {
    local node_name=$1

    for idx in $(yq '.spec.kernelParameters.parameters | keys' $NODECONFIG_FILE | wc -l); do
        local param_name=$(yq ".spec.kernelParameters.parameters[$((idx-1))].name" $NODECONFIG_FILE)
        local expected_value=$(yq ".spec.kernelParameters.parameters[$((idx-1))].value" $NODECONFIG_FILE)

        echo "➖  Checking kernel parameter $param_name on $node_name..."
        # Prepare the command to check kernel parameter
        local check_cmd="sysctl -n '$param_name'"

        local value=$(run_on_node $node_name "$check_cmd" | sed -e 's/\r//g' -e 's/\t/ /g')
        if [ "$value" == "$expected_value" ]; then
            echo "✅  Success: Kernel parameter $param_name is set to $expected_value on $node_name"
        else
            echo "❌  Error: Kernel parameter $param_name is $value on $node_name"
        fi
    done
}

# Function to check host entries on each node
check_hosts() {
    local node_name="$1"

    for idx in $(yq '.spec.hosts.hosts | keys' $NODECONFIG_FILE | wc -l); do
        local hostname=$(yq ".spec.hosts.hosts[$((idx-1))].hostname" $NODECONFIG_FILE)
        local ip=$(yq ".spec.hosts.hosts[$((idx-1))].ip" $NODECONFIG_FILE)

        echo "➖  Checking /etc/hosts on $node_name for entry $ip $hostname..."
        # Prepare the command to check /etc/hosts
        local check_cmd="grep '$ip $hostname' /etc/hosts"

        # Execute the command on the node
        local result=$(run_on_node $node_name "$check_cmd")

        # Check if the output is not empty
        if [[ ! -z "$result" ]]; then
            echo "✅  Success: Host entry $ip $hostname found in /etc/hosts on $node_name"
        else
            echo "❌  Error: Host entry $ip $hostname not found in /etc/hosts on $node_name"
        fi
    done
}

# Function to check kernel modules
check_kernel_modules() {
    local node_name=$1

    for idx in $(yq '.spec.kernelModules.modules | keys' $NODECONFIG_FILE | wc -l); do
        local module=$(yq ".spec.kernelModules.modules[$((idx-1))]" $NODECONFIG_FILE)
        echo "➖  Checking kernel module $module on $node_name..."
        local check_cmd="modinfo $module"
        run_on_node $node_name "$check_cmd" > /dev/null
        local rc=$?
        if [ $rc -eq 0 ]; then
            echo "✅  Success: kernel module $module is loaded on $node_name"
        else
            echo "❌  Error: kernel module $module is not loaded on $node_name"
        fi
    done
}

check_systemd_units() {
    local node_name=$1

    for idx in $(yq '.spec.systemdUnits.units | keys' $NODECONFIG_FILE | wc -l); do
        local name=$(yq ".spec.systemdUnits.units[$((idx-1))].name" $NODECONFIG_FILE)
        echo "➖  Checking systemd unit $name on $node_name..."
        local check_cmd="systemctl --no-pager status nco-$name"
        run_on_node $node_name "$check_cmd" > /dev/null
        local rc=$?
        if [ $rc -eq 0 ]; then
            echo "✅  Success: systemd unit $name is active on $node_name"
        else
            echo "❌  Error: systemd unit $name is not active on $node_name"
        fi
    done
}

check_systemd_overrides() {
    local node_name=$1

    for idx in $(yq '.spec.systemdOverrides.overrides | keys' $NODECONFIG_FILE | wc -l); do
        local serviceName=$(yq ".spec.systemdOverrides.overrides[$((idx-1))].name" $NODECONFIG_FILE)
        local contents=$(yq ".spec.systemdOverrides.overrides[$((idx-1))].file" $NODECONFIG_FILE | tr '\n' ',')
        echo "➖ Checking systemd override on unit $serviceName..."
        local check_cmd="systemctl cat $serviceName | tr '\n' ',' | grep -F '$contents'"
        run_on_node "$node_name" "$check_cmd" > /dev/null
        local rc=$?
        if [ $rc -eq 0 ]; then
            echo "✅  Success: systemd override for $serviceName is active on $node_name"
        else
            echo "❌ Error: systemd override for $serviceName is not active on $node_name"
        fi
    done
}

check_block_in_files() {
    local node_name=$1

    for idx in $(yq '.spec.blockInFiles.blocks | keys' $NODECONFIG_FILE | wc -l); do
        local filename=$(yq ".spec.blockInFiles.blocks[$((idx-1))].filename" $NODECONFIG_FILE)
        local content=$(yq ".spec.blockInFiles.blocks[$((idx-1))].content" $NODECONFIG_FILE | tr '\n' ',')
        echo "➖ Checking block in file on file $filename..."
        local check_cmd="cat $filename | tr '\n' ',' | grep -F '$content'"
        run_on_node "$node_name" "$check_cmd" > /dev/null
        local rc=$?
        if [ $rc -eq 0 ]; then
            echo "✅ Success: block in file $filename is correct on $node_name"
        else
            echo "❌ Error: block in file $filename is not correct on $node_name"
        fi
    done
}

check_certificates() {
    local node_name=$1

    echo "❗ Warning: Checking the certificates module is not possible in Minikube"

    #for idx in $(yq '.spec.certificates.certificates | keys' $NODECONFIG_FILE | wc -l); do
    #    local filename="/etc/ssl/certs/ca-certificates.crt"
    #    local content=$(yq ".spec.certificates.certificates[$((idx-1))].content" $NODECONFIG_FILE | tr '\n' ',')
    #    echo "➖ Checking certificate on file $filename..."
    #    local check_cmd="cat $filename | tr '\n' ',' | grep -F '$content'"
    #    run_on_node "$node_name" "$check_cmd" > /dev/null
    #    local rc=$?
    #    if [ $rc -eq 0 ]; then
    #        echo "✅ Success: certificate in file $filename is correct on $node_name"
    #    else
    #        echo "❌ Error: certificate in file $filename is not correct on $node_name"
    #    fi
    #done
}