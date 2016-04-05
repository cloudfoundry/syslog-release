# syslog-release

A BOSH release to configure rsyslog. This DOES NOT install rsyslog, as it is already included by default in stemcells. It is only used to configure the already included rsyslog instance.

The release will configure rsyslog to use UDP with no encryption by default. If encryption is desired, provide the ca_cert and permitted_peer properties and specify TCP or RELP as transport.
