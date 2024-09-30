# User guide

## Applying configurations

All configurations are declared in NodeConfig manifests, which are applied in
the same way as all kubernetes objects.

1. Write your manifest, for instance:

```yaml
apiVersion: configuration.whitestack.com/v1beta1
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
apiVersion: configuration.whitestack.com/v1beta1
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
