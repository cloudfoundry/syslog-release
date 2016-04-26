# Syslog BOSH Release

This is a BOSH release to forward local syslog events to a remote syslog endpoint. It currently uses [rsyslog](http://www.rsyslog.com/) which is pre-installed by the stemcell.


## Usage

Download the latest release from [bosh.io](https://bosh.io/releases/github.com/cloudfoundry/syslog-release) and include it in your manifest:

```yml
releases:
- name: syslog
  version: latest
```


### Configure Forwarding

Add the [`syslog_forwarder`](https://bosh.io/jobs/syslog_forwarder?source=github.com/cloudfoundry/syslog-release) job to forward all local syslog messages from an instance to a syslog endpoint. Configure `address` and, optionally, `port` and `transport`:

```yml
instance_groups:
- name: some-instance-group
  jobs:
  - name: syslog_forwarder
    release: syslog
  properties:
    syslog:
      address: <IP or hostname>
```

By default, if the syslog endpoint is unavailable messages will be queued. Alternatively, configure `fallback_servers` for higher availability. Only TCP or RELP are supported for fallback functionality:

```yml
properties:
  syslog:
    address: 10.10.10.100
    fallback_servers:
    - address: 10.10.10.101
    - address: 10.10.10.102
```

TLS is supported with additional properties. The following example would forward syslog messages to [papertrailapp.com](https://papertrailapp.com/):

```yml
properties:
  syslog:
    address: logs4.papertrailapp.com
    port: 12345
    transport: tcp
    tls_enabled: true
    permitted_peer: "*.papertrailapp.com"
    ca_cert: |
      -----BEGIN CERTIFICATE-----
      MIIFdDCCBFygAwIBAgIQJ2buVutJ846r13Ci/ITeIjANBgkqhkiG9w0BAQwFADBv
      ...
      pu/xO28QOG8=
      -----END CERTIFICATE-----
      -----BEGIN CERTIFICATE-----
      MIIENjCCAx6gAwIBAgIBATANBgkqhkiG9w0BAQUFADBvMQswCQYDVQQGEwJTRTEU
      ...
      mnkPIAou1Z5jJh5VkpTYghdae9C8x49OhgQ=
      -----END CERTIFICATE-----
```

Note that you may need to include the *entire* certificate chain in `ca_cert` for the forwarding to work. The `openssl` command can be used to view an endpoint's certificate chain:

```bash
openssl s_client -showcerts -servername logs4.papertrailapp.com -connect papertrailapp.com:443 < /dev/null
```


### Test Store

The [`syslog_storer`](https://bosh.io/jobs/syslog_storer?source=github.com/cloudfoundry/syslog-release) is meant for testing. Deploy it and configure your instances to forward logs to it. It should not be co-located with other jobs which also try to configure syslog. Received logs are stored in `/var/vcap/store/syslog_storer/syslog.log`.

```yml
instance_groups:
- name: syslog_storer
  jobs:
  - name: syslog_storer
    release: syslog
```

Remember to allow inbound traffic on TCP port 514 in your IaaS security groups.


## Tech Notes

RSYSLOG is system for log processing; it is a drop-in replacement for the UNIX's venerable [syslog](https://en.wikipedia.org/wiki/Syslog), which logs messages to various files and/or log hosts. RSYSLOG can be configured as a **storer** (i.e. it receives log messages from other hosts) or a **forwarder** (i.e. it forwards system log messages to RSYSLOG storers, syslog servers, or log aggregation services).

The RSYSLOG configuration file is `/etc/rsyslog.conf`. The RSYSLOG forwarder's customizations are rendered into `/etc/rsyslog.d/rsyslog.conf`, which is included by the configuration file.


## License

[Apache License Version 2.0](./LICENSE)
