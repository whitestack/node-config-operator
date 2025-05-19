# User guide

## Applying configurations

All configurations are declared in NodeConfig manifests, which are applied in
the same way as all kubernetes objects.

1. Write your manifest, for instance:

```yaml
apiVersion: configuration.whitestack.com/v1beta2
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  kernelParameters:
    parameters:
    - name: fs.file-max
      value: "54321"
    state: present
```

1. Save it with a convenient name, e.g. `sample_node_config.yaml`

1. In order to apply the manifest, execute the following command:

    `kubectl apply -f sample_node_config.yaml`

## Grouping configurations with node selectors

NodeConfig objects can be limited to specific nodes by using kubernetes labels.

Consider a cluster with nodes node-0, node-1 and node-2, and the previous
`NodeConfig` manifest.

In order to apply the configuration to node-0 and node-1 only, execute the
following procedure:

1. Add a label to nodes `node-0` and `node-1`:
   `kubectl label node-0 node-1 mylabel=test`

1. Edit the `NodeConfig` object and add a `nodeSelector` with the same label:

```yaml
apiVersion: configuration.whitestack.com/v1beta2
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  nodeSelector:
  - key: "mylabel"
    operator: In
    values:
    - "test"
  kernelParameters:
    parameters:
    - name: fs.file-max
      value: "54321"
    state: present
```

1. Apply the changes.

## Removing configurations

To remove a module's configuration you have to set the `state` field to `absent`
and apply the CR to the cluster. You can check in the logs if it's been
successfully removed.

## Configuration

In the helm chart you have these options to configure the `NodeConfig` operator:

- `aptEnabled`: the [`apt` module](/docs/module_reference.md#apt-packages)
  requires this flag to be set. It also schedules a job to update the apt
  package cache every 5 hours.
- `hostfsEnabled`: this flag mounts the host's root filesystem in the controller
  pod. This flag is required for [some modules][modules].
- `validationModulePresentEnabled`: this flag enables the validation that checks
  if a module is defined multiple times for the same node selector in the
  validation webhook.
- `ignoreNodeReady`: by default the controller will not reconcile a resource if
  the node is `NotReady` but you can ignore that check by setting this flag to
  true.

[modules]: ./module_reference.md
