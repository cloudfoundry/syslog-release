# RSYSLOG BOSH Release

This is a BOSH release to deploy RSYSLOG. This release does *not* install RSYSLOG (it is already included by default in stemcells), it merely configures it.

### Upload Release to BOSH Director

Clone this repo, create the release, and upload the release to your BOSH director:
```bash
git clone https://github.com/cloudfoundry/syslog-release.git
cd syslog-release
bosh create release --force
bosh upload release
```

### Create RSYSLOG Server

This is how to create an RSYSLOG server which receives
syslog messages on UDP port 514 (the default). The RSYSLOG server job can be co-located with other jobs (e.g. Redis).

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

### Configure an *instance_group* to forward syslog messages to an RSYSLOG server.

This is how to configure an instance_group to forward syslog messages
to the RSYSLOG server on UDP port 514 (the default).

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
       destination_address: <RSYSLOG server's IP address or fully-qualified domain name>
       destination_transport: udp
       destination_port: 514
    ```

The release will configure rsyslog to use UDP with no encryption by default. If encryption is desired, provide the ca_cert and permitted_peer properties and specify TCP or RELP as transport.

### RSYSLOG Failover configuration

In the event of a failure of a log server, the RSYSLOG forwarder instance group can be configured to forward syslog messages to a failover server. Note that the transport cannot be UDP (lossy protocol) for failover to work, but TCP and RELP will work.

In this example, we configure our primary log server to be 10.10.10.100, and our failover server to be 10.10.10.99:

```yml
properties:
  destination_address: 10.10.10.100
  destination_port: 514
  destination_transport: tcp
  fallback_addresses:
  - address: 10.10.10.99
    port: 514
    transport: tcp
```

### Tech Notes

The RSYSLOG server stores its syslog messages in `/var/vcap/store/syslog_storer/syslog.log`.

The RSYSLOG configuration file is `/etc/rsyslog.conf`. The RSYSLOG forwarder's customizations are rendered into `/etc/rsyslog.d/rsyslog.conf`, which is included by the configuration file.
