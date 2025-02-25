# API Reference

## Packages
- [configuration.whitestack.com/v1beta1](#configurationwhitestackcomv1beta1)
- [configuration.whitestack.com/v1beta2](#configurationwhitestackcomv1beta2)


## configuration.whitestack.com/v1beta1

Package v1 contains API Schema definitions for the configuration v1 API group

### Resource Types
- [NodeConfig](#nodeconfig)
- [NodeConfigList](#nodeconfiglist)



#### NodeConfig



NodeConfig is the Schema for the nodeconfigs API

_Appears in:_
- [NodeConfigList](#nodeconfiglist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `configuration.whitestack.com/v1beta1`
| `kind` _string_ | `NodeConfig`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[NodeConfigSpec](#nodeconfigspec)_ |  |


#### NodeConfigList



NodeConfigList contains a list of NodeConfig



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `configuration.whitestack.com/v1beta1`
| `kind` _string_ | `NodeConfigList`
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` _[NodeConfig](#nodeconfig) array_ |  |


#### NodeConfigSpec



NodeConfigSpec defines the desired state of NodeConfig

_Appears in:_
- [NodeConfig](#nodeconfig)

| Field | Description |
| --- | --- |
| `kernelParameters` _[KernelParameters](#kernelparameters)_ | List of kernel parameters (sysctl). Each parameter should contain name and value |
| `kernelModules` _[KernelModules](#kernelmodules)_ | List of kernel modules to load |
| `systemdUnits` _[SystemdUnits](#systemdunits)_ | List of systemd units to install |
| `systemdOverrides` _[SystemdOverrides](#systemdoverrides)_ | List of systemd overrides to add to existing systemd units |
| `hosts` _[Hosts](#hosts)_ | List of hosts to install to /etc/hosts |
| `aptPackages` _[AptPackages](#aptpackages)_ | List of apt packages to install |
| `blockInFiles` _[BlockInFiles](#blockinfiles)_ | List of blocks to add to files |
| `nodeSelector` _[LabelSelectorRequirement](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#labelselectorrequirement-v1-meta) array_ | Defines the target nodes for this NodeConfig (optional, default is apply to all nodes) |





## configuration.whitestack.com/v1beta2

Package v1beta2 contains API Schema definitions for the configuration v1beta2 API group

### Resource Types
- [NodeConfig](#nodeconfig)
- [NodeConfigList](#nodeconfiglist)



#### NodeConfig



NodeConfig is the Schema for the nodeconfigs API

_Appears in:_
- [NodeConfigList](#nodeconfiglist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `configuration.whitestack.com/v1beta2`
| `kind` _string_ | `NodeConfig`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[NodeConfigSpec](#nodeconfigspec)_ |  |


#### NodeConfigList



NodeConfigList contains a list of NodeConfig



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `configuration.whitestack.com/v1beta2`
| `kind` _string_ | `NodeConfigList`
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` _[NodeConfig](#nodeconfig) array_ |  |


#### NodeConfigSpec



NodeConfigSpec defines the desired state of NodeConfig

_Appears in:_
- [NodeConfig](#nodeconfig)

| Field | Description |
| --- | --- |
| `kernelParameters` _[KernelParameters](#kernelparameters)_ | List of kernel parameters (sysctl). Each parameter should contain name and value |
| `kernelModules` _[KernelModules](#kernelmodules)_ | List of kernel modules to load |
| `systemdUnits` _[SystemdUnits](#systemdunits)_ | List of systemd units to install |
| `systemdOverrides` _[SystemdOverrides](#systemdoverrides)_ | List of systemd overrides to add to existing systemd units |
| `hosts` _[Hosts](#hosts)_ | List of hosts to install to /etc/hosts |
| `aptPackages` _[AptPackages](#aptpackages)_ | List of apt packages to install |
| `blockInFiles` _[BlockInFiles](#blockinfiles)_ | List of blocks to add to files |
| `certificates` _[Certificates](#certificates)_ | List of Certificates to add to /etc/ssl/certs |
| `crontabs` _[Crontabs](#crontabs)_ | List of Crontabs to schedule |
| `grubKernelConfig` _[GrubKernel](#grubkernel)_ | GrubKernelConfig contains kernel version and command line arguments for GRUB configuration |
| `nodeSelector` _[LabelSelectorRequirement](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#labelselectorrequirement-v1-meta) array_ | Defines the target nodes for this NodeConfig (optional, default is apply to all nodes) |




