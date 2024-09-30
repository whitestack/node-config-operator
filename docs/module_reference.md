# Module reference and examples

## Kernel parameter

Kernel parameters are configured using `sysctl` commands. For this example, you
need to configure the following kernel parameters in the kubernetes nodes:

`sysctl -w kernel.hostname=test-hostname`

This configuration can be applied with the following CR:

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

## Etc hosts

Hostname resolution can be configured locally in `/etc/hosts`. For this example,
you need to add the following entries:

```bash
test.whitestack.com 10.0.0.1
test2.whitestack.com 10.0.0.2
test3.whitestack.com test4.whitestack.com 10.0.0.3
```

This configuration can be applied with the following CR:

```yaml
apiVersion: configuration.whitestack.com/v1beta1
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  hosts:
    hosts:
    - hostname: "test.whitestack.com"
      ip: "10.0.0.1"
    - hostname: "test2.whitestack.com"
      ip: "10.0.0.2"
    - hostname: "test3.whitestack.com test4.whitestack.com"
      ip: "10.0.0.3"
    state: present
```

## Systemd units

> [!NOTE]
> This module requires that the `managerConfig.hostfsEnabled` option is set to
> true

New systemd units can be configured in `/etc/systemd/system`. For this example,
you need to configure a service with name `test-service` and the following unit
file:

```systemd
[Unit]
Description=example systemd service unit file.

[Service]
ExecStart=/bin/bash /usr/sbin/example.sh

[Install]
WantedBy=multi-user.target
```

This configuration can be applied with the following CR:

```yaml
apiVersion: configuration.whitestack.com/v1beta1
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  systemdUnits:
    units:
    - name: test-service
      file: |
        [Unit]
        Description=example systemd service unit file.

        [Service]
        ExecStart=/bin/bash /usr/sbin/example.sh

        [Install]
        WantedBy=multi-user.target
    state: present
```

## Systemd overrides

> [!NOTE]
> This module requires that the `managerConfig.hostfsEnabled` option is set to
> true

Systemd overrides will be written to `/etc/systemd/system/<unit-name>.d`. For
example, if you want to change the `Exec` command to a running service you use
the following CR:

```yaml
apiVersion: configuration.whitestack.com/v1beta1
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  systemdOverrides:
    overrides:
    - name: getty@tty2.service
      file: |
        [Service]
        ExecStart=
        ExecStart=sleep 2000
    state: present
```

## Kernel modules

Kernel modules can be loaded with `modprobe`. For this example, you need to load
the following kernel modules:

```bash
modprobe multipath
modprobe dm_multipath
```

This configuration can be applied with the following CR:

```yaml
apiVersion: configuration.whitestack.com/v1beta1
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  kernelModules:
    modules:
    - multipath
    - dm_multipath
    state: present
```

## Block in File

> [!NOTE]
> This module requires that the `managerConfig.hostfsEnabled` option is set to
> true

This module adds a block in a new or existing file. For example for a file in
`/etc/test.test`:

```text
line0
line1
```

You can add a block to the end of the file:

```text
line0
line1
# BEGIN MARKER
line11
line12
# END MARKER
```

With this CR:

```yaml
apiVersion: configuration.whitestack.com/v1beta1
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  blockInFiles:
    blocks:
    - filename: "/etc/test.test"
      content: |
        line11
        line12
      beginMarker: "# BEGIN MARKER"
      endMarker: "# END MARKER"
    state: present
```

## Apt packages

> [!NOTE]
> This module requires that the `managerConfig.hostfsEnabled` and
> `managerConfig.aptEnabled` options are set to true. Only works in Ubuntu
> servers

This module updates the apt's package lists and installs the packages defined in
the CR, for example:

```yaml
apiVersion: configuration.whitestack.com/v1beta1
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  aptPackages:
    packages:
    - name: vim
    - name: ssh
      version: 1:9.6p1-3ubuntu13.5
    state: present
```

Will install the latest version of the `vim` package and the required version of
the `ssh` package.
