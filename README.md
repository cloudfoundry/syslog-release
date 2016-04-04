# syslog-release

A BOSH release to configure rsyslog. This DOES NOT install rsyslog, as it is already included by default in stemcells. It is only used to configure the already included rsyslog instance.

The release will configure rsyslog to use UDP with no encryption by default. If encryption is desired, provide the ca_cert and permitted_peer properties and specify TCP or RELP as transport.

## Properties

* destination: IP or DNS address of the syslog drain
* port: Port of the syslog drain
* transport: One of ["udp", "tcp", "relp"]. Defaults to "udp".
* custom_rule: Custom rule for syslog event forwarder
* permitted_peer: Accepted fingerprint (SHA1) or name of remote peer. Required if TLS is enabled
* ca_cert: Trust these Certificate Authorities
