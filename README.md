# RSYSLOG BOSH Release

This is a BOSH release of [RSYSLOG](http://www.rsyslog.com/). This release does *not* install RSYSLOG (it is already included by default in stemcells), it merely configures it.

RSYSLOG is system for log processing; it is a drop-in replacement for the UNIX's venerable
[syslog](https://en.wikipedia.org/wiki/Syslog), which logs messages to various files and/or log hosts.
RSYSLOG can be configured as a **storer** (i.e. it receives log messages from other hosts)
or a **forwarder** (i.e. it forwards system log messages to RSYSLOG storers, syslog servers, or log aggregation services).

### Create RSYSLOG Storer

This is how to create an RSYSLOG storer which receives syslog messages on UDP port 514 (the default). The RSYSLOG storer job can be co-located with other jobs (e.g. Redis). This job is not meant to be used for a production storage of logs, but rather is included here for testing.

1. Include `syslog-release` in the `releases` section of the deployment manifest

  ```yml
  releases:
  - name: syslog-release
    version: latest
  ```
2. Create an `instance_group` with a `job` that has the `syslog-release`
  ```yml
  instance_groups:
  - name: syslog_storer
    jobs:
    - name: syslog_storer
      release: syslog-release
    properties:
      transport: udp
      port: 514
  ```

3. Deploy
  ```bash
  bosh deploy
  ```

Make sure that any packet filter (e.g. Amazon AWS security groups) allow inbound traffic on UDP port 514.

### Create an RSYSLOG Forwarder

This is how to configure an instance_group to forward syslog messages to the RSYSLOG storer on UDP port 514 (the default). Note that RSYSLOG Forwarders are almost always co-located with other jobs.

1. Include `syslog-release` in the `releases` section of the deployment manifest

  ```yml
  releases:
  - name: syslog-release
    version: latest
  ```
2. Configure deployment manifest

   ```yml
   instance_groups:
   - name: some-instance-group
     jobs:
     - name: syslog_forwarder
       release: syslog-release
     properties:
       address: <RSYSLOG storer's IP address or fully-qualified domain name>
       transport: udp
       port: 514
    ```

### Create an RSYSLOG Forwarder with Failover

In the event of a failure of a log storer, the RSYSLOG forwarder instance group can be configured to forward syslog messages to a failover storer. Failover requires the use of a lossless transport (i.e. TCP or RELP); failover will not work with UDP.

In this example, we configure our primary log storer to be 10.10.10.100, and our failover storer to be 10.10.10.99:

```yml
properties:
  address: 10.10.10.100
  port: 514
  transport: tcp
  fallback_servers:
  - address: 10.10.10.99
    port: 514
    transport: tcp
```

### Create an RSYSLOG Forwarder with TLS (Encryption)

In this example, we configure our RSYSLOG to forward syslog messages to papertrailapp.com,
a popular log aggregation service. For brevity we truncated the SSL certificates; note that you must include the *entire* certificate chain for the forwarding to work. Also `port` will be different for your *papertrail* account.

```yml
properties:
  address: logs4.papertrailapp.com
  port: 41120
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

### Tech Notes

The RSYSLOG storer stores its syslog messages in `/var/vcap/store/syslog_storer/syslog.log`.

The RSYSLOG configuration file is `/etc/rsyslog.conf`. The RSYSLOG forwarder's customizations are rendered into `/etc/rsyslog.d/rsyslog.conf`, which is included by the configuration file.

To configure RSYSLOG to use TLS, you must populate the `ca_cert` section of the job's
properties section with a valid
certificate chain.
Use the following command to extract the certificate chain from the papertrailapp.com webserver.

```bash
openssl s_client -showcerts -servername logs4.papertrailapp.com -connect papertrailapp.com:443 < /dev/null
```
