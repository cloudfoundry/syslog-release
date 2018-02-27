# On Security Events
When syslog forwarding is enabled,
linux system logs are forwarded
along with logs from BOSH jobs.

The [stemcell][stemcell-builder] includes opinionated configuration
on top of the default distribution configuration coming from Canonical.
This configuration includes [auditd][auditd-man] and friends,
which means the system logs include system security event audit information.
When we say "system security events,"
we mean things like SSH connections,
user creation,
and permissions changes,
just to name a few.

Here, we'll provide some example security events.
We'll describe how to provoke them,
and how to validate that they were logged.
Our hope is that this will be useful
to anyone evaluating the audit trail features of the Cloud Foundry platform.

Note that while security-event-related-logs emitted by BOSH jobs
will also be forwarded,
they're not what we're aiming to discuss here.

## Examples

[auditd-man]: http://manpages.ubuntu.com/manpages/trusty/man8/auditd.8.html
[stemcell-builder]: https://github.com/cloudfoundry/bosh-linux-stemcell-builder
