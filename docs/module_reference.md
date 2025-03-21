# Module reference and examples

## Kernel parameter

Kernel parameters are configured using `sysctl` commands. For this example, you
need to configure the following kernel parameters in the kubernetes nodes:

`sysctl -w kernel.hostname=test-hostname`

This configuration can be applied with the following CR:

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
    priority: 90
```

You can add an optional `priority` key to set the priority for these parameters.
Default priority is 50

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
apiVersion: configuration.whitestack.com/v1beta2
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
apiVersion: configuration.whitestack.com/v1beta2
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
apiVersion: configuration.whitestack.com/v1beta2
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
      priority: 20
    state: present
```

You can add an optional `priority` key to set the priority for each override.
Default priority is 50

## Kernel modules

Kernel modules can be loaded with `modprobe`. For this example, you need to load
the following kernel modules:

```bash
modprobe multipath
modprobe dm_multipath
```

This configuration can be applied with the following CR:

```yaml
apiVersion: configuration.whitestack.com/v1beta2
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  kernelModules:
    modules:
    - multipath
    - dm_multipath
    state: present
    priority: 50
```

You can add an optional `priority` key to set the priority for these modules.
Default priority is 50

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
apiVersion: configuration.whitestack.com/v1beta2
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

## Certificates

> [!NOTE]
> This module requires that the `managerConfig.hostfsEnabled` option is set to
> true

This module adds a text to a new or existing certificate file. For example for a
file in `/etc/ssl/certs/whitestack.crt`:

With this CR:

```yaml
apiVersion: configuration.whitestack.com/v1beta2
kind: NodeConfig
metadata:
  name: nodeconfig-sample
spec:
  certificates:
    certificates:
    - filename: "whitestack.crt"
      content: |
        -----BEGIN CERTIFICATE-----
        MIIE0DCCA7igAwIBAgIBBzANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMCVVMx
        EDAOBgNVBAgTB0FyaXpvbmExEzARBgNVBAcTClNjb3R0c2RhbGUxGjAYBgNVBAoT
        EUdvRGFkZHkuY29tLCBJbmMuMTEwLwYDVQQDEyhHbyBEYWRkeSBSb290IENlcnRp
        ZmljYXRlIEF1dGhvcml0eSAtIEcyMB4XDTExMDUwMzA3MDAwMFoXDTMxMDUwMzA3
        MDAwMFowgbQxCzAJBgNVBAYTAlVTMRAwDgYDVQQIEwdBcml6b25hMRMwEQYDVQQH
        EwpTY290dHNkYWxlMRowGAYDVQQKExFHb0RhZGR5LmNvbSwgSW5jLjEtMCsGA1UE
        CxMkaHR0cDovL2NlcnRzLmdvZGFkZHkuY29tL3JlcG9zaXRvcnkvMTMwMQYDVQQD
        EypHbyBEYWRkeSBTZWN1cmUgQ2VydGlmaWNhdGUgQXV0aG9yaXR5IC0gRzIwggEi
        MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC54MsQ1K92vdSTYuswZLiBCGzD
        BNliF44v/z5lz4/OYuY8UhzaFkVLVat4a2ODYpDOD2lsmcgaFItMzEUz6ojcnqOv
        K/6AYZ15V8TPLvQ/MDxdR/yaFrzDN5ZBUY4RS1T4KL7QjL7wMDge87Am+GZHY23e
        cSZHjzhHU9FGHbTj3ADqRay9vHHZqm8A29vNMDp5T19MR/gd71vCxJ1gO7GyQ5HY
        pDNO6rPWJ0+tJYqlxvTV0KaudAVkV4i1RFXULSo6Pvi4vekyCgKUZMQWOlDxSq7n
        eTOvDCAHf+jfBDnCaQJsY1L6d8EbyHSHyLmTGFBUNUtpTrw700kuH9zB0lL7AgMB
        AAGjggEaMIIBFjAPBgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQEAwIBBjAdBgNV
        HQ4EFgQUQMK9J47MNIMwojPX+2yz8LQsgM4wHwYDVR0jBBgwFoAUOpqFBxBnKLbv
        9r0FQW4gwZTaD94wNAYIKwYBBQUHAQEEKDAmMCQGCCsGAQUFBzABhhhodHRwOi8v
        b2NzcC5nb2RhZGR5LmNvbS8wNQYDVR0fBC4wLDAqoCigJoYkaHR0cDovL2NybC5n
        b2RhZGR5LmNvbS9nZHJvb3QtZzIuY3JsMEYGA1UdIAQ/MD0wOwYEVR0gADAzMDEG
        CCsGAQUFBwIBFiVodHRwczovL2NlcnRzLmdvZGFkZHkuY29tL3JlcG9zaXRvcnkv
        MA0GCSqGSIb3DQEBCwUAA4IBAQAIfmyTEMg4uJapkEv/oV9PBO9sPpyIBslQj6Zz
        91cxG7685C/b+LrTW+C05+Z5Yg4MotdqY3MxtfWoSKQ7CC2iXZDXtHwlTxFWMMS2
        RJ17LJ3lXubvDGGqv+QqG+6EnriDfcFDzkSnE3ANkR/0yBOtg2DZ2HKocyQetawi
        DsoXiWJYRBuriSUBAA/NxBti21G00w9RKpv0vHP8ds42pM3Z2Czqrpv1KrKQ0U11
        GIo/ikGQI31bS/6kA1ibRrLDYGCD+H1QQc7CoZDDu+8CL9IVVO5EFdkKrqeKM+2x
        LXY2JtwE65/3YR8V3Idv7kaWKK2hJn0KCacuBKONvPi8BDAB
        -----END CERTIFICATE-----
    state: present
```

You may include multiple certificates in a single file, and multiple files in a
single custom resource.

## Apt packages

> [!NOTE]
> This module requires that the `managerConfig.hostfsEnabled` and
> `managerConfig.aptEnabled` options are set to true. Only works in Ubuntu
> servers

This module updates the apt's package lists and installs the packages defined in
the CR, for example:

```yaml
apiVersion: configuration.whitestack.com/v1beta2
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

## Crontabs

Crontab entries can be managed by creating or removing files in the
`/etc/cron.d` directory. Each entry specifies a scheduled task to be executed
under a specific user. For example, to schedule a daily backup script to run as
the `root` user, you would create a crontab entry like this:

```shell
@daily root /usr/bin/backup.sh # daily-backup
```

This configuration can be applied using the following Custom Resource (CR):

```yaml
apiVersion: configuration.whitestack.com/v1beta2
kind: NodeConfig
metadata:
  name: nodeconfig-crontab-sample
spec:
  crontabs:
    entries:
    - name: "daily-backup"
      special_time: "daily"
      job: "/usr/bin/backup.sh"
      user: "root"
    - name: "hourly-cleanup"
      minute: "0"
      hour: "*"
      dayOfMonth: "*"
      month: "*"
      dayOfWeek: "*"
      job: "/usr/bin/cleanup.sh"
      user: "root"
    state: "present"
```

Fiels:

- name: A unique identifier for the cron job. This is used to generate the
  filename in `/etc/cron.d`.
- special_time: (Optional) Specifies a predefined schedule such as `@daily`,
  `@reboot`, `@weekly`, etc.
- minute , hour , dayOfMonth , month , dayOfWeek: (Optional) Define the schedule
  explicitly. Defaults to `*` if not specified.
- job: The command or script to execute.
- user: The user under which the task will run.

## GRUB Kernel Config

The GRUB configuration can be managed by creating or removing files in the
`/etc/default/grub.d` directory. This approach enables modular and idempotent
management of GRUB settings, such as kernel command-line arguments and the
default kernel version.

For example, to set specific kernel arguments and select a default kernel
version, you would create a configuration file like this:

```shell
# BEGIN MARKER NCO GRUB CONFIG
GRUB_CMDLINE_LINUX="quiet splash"
GRUB_DEFAULT="Advanced options for Ubuntu>Ubuntu, with Linux 5.15.0-91-generic"
# END MARKER NCO GRUB CONFIG
```

This configuration can be applied using the following Custom Resource (CR):

```yaml
apiVersion: configuration.whitestack.com/v1beta2
kind: NodeConfig
metadata:
  name: nodeconfig-grub-sample
spec:
  grubKernelConfig:
    kernelVersion: "5.15.0-91-generic"
    args:
      - "quiet"
      - "splash"
    state: "present"
    priority: 55
```

You can add an optional `priority` key to set the priority for this
configuration. Default priority is 50

Fields:

- kernelVersion: (Optional) Specifies the Linux kernel version to set as the
  default (e.g., "5.15.0-91-generic"). If not provided, the default kernel
  will remain unchanged.
- args: (Optional) A list of kernel command-line arguments to be added to
  `GRUB_CMDLINE_LINUX`. If not specified, no changes will be made to the
  kernel command-line arguments.
