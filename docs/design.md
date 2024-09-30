# Design

## Concept

There are use-cases where the Kubernetes administrator needs to configure the
nodes of a cluster, such as loading kernel modules, changing the kernel
configuration, adding or modifying systemd units, etc. The current solution is
to use Ansible but it has some drawbacks:

1. It requires an external system to store and apply the required state.
1. The node state can drift from the requested state if it's not applied
   periodically.

With this operator we try to solve both of those problems by running an agent in
each node as a pod and storing the requested state inside Kubernetes as
CustomResources.

## Modules

The available modules are:

1. apt: installs apt packages
1. kernel modules: loads kernel modules
1. kernel parameters: changes kernel configuration via sysctl
1. hosts: adds entries to `/etc/hosts`
1. block-in-file: adds a block of text in a new or existing file
1. systemd: adds and starts a new systemd service
1. systemd-override: adds an override for an existing systemd unit

And they're applied in this order.

The modules:

- apt
- block-in-file
- systemd
- systemd-override

Require that the Helm value `managerConfig.hostfsEnabled` is set to true as they
need to mount the whole host filesystem to the pod so they can run executables
in the root namespace. Additionally, the apt module requires tha the value
`managerConfig.aptEnabled` is set to true to enable an internal cron that
periodically updates the apt package list.

All modules have a `state` field that indicates whether the module's
configuration will be applied or removed from the node. Possible values for this
field are `present` or `absent`.

## Deployment in Kubernetes

This operator is deployed as a DaemonSet in Kubernetes so that each node in the
cluster can be configured via our CustomResource. The operator

Each pod watches all the `NodeConfig` CRs in the cluster and runs each module's
reconciliation loop. When all the modules are reconciled, each pod will add a
node finalizer to the CR so it can be cleaned up when the CR is deleted.
