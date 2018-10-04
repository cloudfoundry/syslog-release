# Syslog BOSH Release
# This has been modified to forward pg_log prefix. 
* Slack: #syslog on <https://slack.cloudfoundry.org>
* Tracker: [CF Platform Logging Improvements][tracker]
* CI Pipelines: https://syslog.ci.cf-app.com

1. [Usage](#usage)
1. [Configuration](#configure-forwarding)
1. [Custom Rules](#custom-rule)
1. [Output Format](#format)
1. [Tech Notes](#tech-notes)
1. [Development](#development)
1. [CI](#CI)

This is a BOSH release
to forward local syslog events
in [RFC5424][RFC] format
to a remote syslog endpoint.
It currently uses [RSYSLOG](http://www.rsyslog.com/)
which is pre-installed by the stemcell.

There is a [separate release][windows-syslog]
to accomplish this on Windows stemcells,
which uses blackbox, but not rsyslog.

## Usage
If you looking for the original release.
You can find it from [bosh.io][syslog-bosh-io]
and include it in your manifest:

```yml
releases:
- name: syslog
  version: latest
```

If you are deploying the Cloud Foundry Application Runtime
using [`cf-deployment`][cf-d],
there is an [ops-file][syslog-addon-ops] available
that will add the syslog release and syslog_forwarder job,
and expose configuration variables.

Otherwise, you can co-locate
and configure
the `syslog_forwarder` yourself.

### Configure Forwarding
Add the [`syslog_forwarder`][forwarder-spec-page]
to forward all local syslog messages
from an instance
to a syslog endpoint.
Configure `address` and,
optionally,
`port` and `transport`:

### Configure Log_prfix
Add the [`log_suffix`]
to allow a different log suffix to be forwarded by blackbox.
The default value is "log". 
Which will forwarding both system and postgresql logs. 

```yml
instance_groups:
- name: some-instance-group
  jobs:
  - name: syslog_forwarder
    release: syslog
  properties:
    syslog:
      address: <IP or hostname>
      blackbox:
        log_suffix: <log_suffix>
```

By default,
if the syslog endpoint is unavailable,
messages will be queued.
Alternatively, configure `fallback_servers`
for higher availability.
Only TCP or RELP are supported
for fallback functionality:

```yml
properties:
  syslog:
    address: 10.10.10.100
    fallback_servers:
    - address: 10.10.10.101
    - address: 10.10.10.102
```

TLS is supported
with additional properties.
The following example
would forward syslog messages
to [papertrail](https://papertrailapp.com/):

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

Alternatively, if the intended syslog recipient's certificate
is signed by any Certificate Authority
in the BOSH instances' cert store
(most common CAs are included on the stemcell),
you can omit the `ca_cert` field entirely.

If you do include `ca_cert`,
please note that the standard
cert store will no longer be referenced.
This necessitates including
the _entire_ certificate chain.

### Custom Rule
This release allows a custom rule
to be inserted before the rule
that accomplishes log forwarding.
This can be useful if you only wish
to forward certain logs,
or if there is a certain type of log
you wish to exclude from forwarding.

We have some simple documentation
with a few example rules in
[`example-custom-rules.md`](examples/example-custom-rules.md).

**Please note:** if your custom rule is invalid,
it will be logged and discarded.

### Test Store
Has been removed for ops environment.
The [`syslog_storer`][storer-spec-page] is meant for testing.
Deploy it and configure your instances to forward logs to it.
It provides a link that the forwarder consumes,
so if they are deployed together and the forwarder is otherwise unconfigured,
logs should be sent to the storer.
It should not be co-located
with other jobs which also try to configure syslog.
Received logs are stored in `/var/vcap/store/syslog_storer/syslog.log`.

You can add it to a deployment manifest
very simply:

```yml
instance_groups:
- name: syslog_storer
  jobs:
  - name: syslog_storer
    release: syslog
```

Remember to allow inbound traffic
on TCP port 514
in your IaaS security groups.

The storer can also be used to test TLS connections.
If you provide a Certificate Authority to the `syslog.tls.generation` properties,
each storer instance will generate a cert signed by that CA at startup,
with the instance's IP address as the common name.
You will need to explicitly configure this CA's cert as trusted on the forwarder.

## Format
This release forwards messages
using the [RFC5424][RFC] standard,
which is natively supported by most log platforms.

Forwarded messages are annotated with structured data.
The [Structured Data ID][sd-id] is `instance@47450`,
which is intended to allow parsing rules specific
to the structured data emitted by BOSH instances
using this release.
(The `47450` is our
[private enterprise number][ent-nums].)
The structured data contains the following fields:

- director
- deployment
- availability zone
- instance group
- instance ID

The whole thing looks something like this:
```
<$PRI>$VERSION $TIMESTAMP $HOST $APP_NAME $PROC_ID $MSG_ID [instance@47450 director="$DIRECTOR" deployment="$DEPLOYMENT" group="$INSTANCE_GROUP" az="$AVAILABILITY_ZONE" id="$ID"] $MESSAGE
```

Here are a couple of example messages from diego:

```
<14>1 2017-01-25T13:25:03.18377Z 192.0.2.10 etcd rs2 - [instance@47450 director="test-env" deployment="cf" group="diego_database" az="us-west1-a" id="83bd66e5-3fdf-44b7-bdd6-508deae7c786"] [INFO] the leader is [https://diego-database-0.etcd.service.cf.internal:4001]
<14>1 2017-01-25T13:25:03.184491Z 192.0.2.10 bbs rs2 - [instance@47450 director="test-env" deployment="cf" group="diego_database" az="us-west1-a" id="83bd66e5-3fdf-44b7-bdd6-508deae7c786"] {"timestamp":"1485350702.539694548","source":"bbs","message":"bbs.request.start-actual-lrp.starting","log_level":1,"data":{"actual_lrp_instance_key":{"instance_guid":"271f71c7-4119-4490-619f-4f44694717c0","cell_id":"diego_cell-2-41f21178-d619-4976-901c-325bc2d0d11d"},"actual_lrp_key":{"process_guid":"1545ce89-01e6-4b8f-9cb1-5654a3ecae10-137e7eb4-12de-457d-8e3e-1258e5a74687","index":5,"domain":"cf-apps"},"method":"POST","net_info":{"address":"192.0.2.12","ports":[{"container_port":8080,"host_port":61532},{"container_port":2222,"host_port":61533}]},"request":"/v1/actual_lrps/start","session":"418.1"}}
```
Note: the `rs2` PROC_ID in the above indicates that the logs
have been forwarded from a file by blackbox,
which uses remote_syslog2 under the covers.

A sample logstash config with filters to extract instance metadata
is in [`scripts/logstash-filters.conf`](examples/logstash-filters.conf).

## Tech Notes
RSYSLOG is a system for log processing;
it is a drop-in replacement for the UNIX's venerable [syslog](https://en.wikipedia.org/wiki/Syslog).
RSYSLOG can be configured as a **storer**
(i.e. it receives log messages from other hosts)
or a **forwarder**
(i.e. it forwards system log messages
to RSYSLOG storers, syslog servers, or log aggregation services).

The default RSYSLOG configuration file is `/etc/rsyslog.conf`.
On the stemcell, this specifies that configuration in`/etc/rsyslog.d/*`
will also be respected.
The RSYSLOG forwarder's customizations
are rendered into several files following the format
`/etc/rsyslog.d/[0-9][0-9]-syslog-release-*.conf`.

**Note:** `syslog-release` deletes files in its pattern,
and `/etc/rsyslog.d/rsyslog.conf`, a legacy config location,
during prestart.

BLACKBOX is used for forwarding from files.
It was historically an experiment by the Concourse team,
but now it's just used to forward logs from files.
It currently vendors a mutated version of
the remote-syslog2 library.
The mutation is to enable filtering of logs
forwarded by blackbox,
and to allow us to support structured data in
windows-syslog-release,
which is purely reliant on blackbox.
At some point it might be worthwhile
to PR those changes back up;
the current situation is admittedly odd and awkward.

RSYSLOG itself might be able to do this at some point in the future.
The current version of rsyslog does not sufficiently support wildcards
for our use-case, but it may be worth further exploring in the future.

## Development
# Skip the submodule update. This will overwrite the modified blackbox.
In order to build releases or run tests,
you will need to initialize and update the blackbox submodule:
```
git submodule init && git submodule update
```

There's a suite of acceptance tests
in the `tests/` directory.
To use them, you will need to [install Go][go-installation].

Before running tests, you will need to create a bosh director.
First you should ensure the bosh2 cli is installed and the bosh-deployment
repository is downloaded and located at `~/workspace/bosh-deployment`. You can then
run `./scripts/setup-bosh-lite-for-tests.sh` to create the director.
Afterwards execute `source export-bosh-lite-creds.sh` to target the bosh director.

(Alternatively, if you wish to run the specs against the CI env,
you can `source` the `.envrc` in your env-repo.)

The tests can then be run from the top of the repo with
`./scripts/test`.

For more details, see [`tests/README.md`][test-readme].

We are unlikely to merge PRs that add features without tests. Please submit all
pull requests against the develop branch.

## CI
Our CI pipelines can be found at https://syslog.ci.cf-app.com.
Pipeline configuration can be found in the `.concourse` directory of this repo.
While our pipeline is principally built using [`cf-deployment-concourse-tasks`][cf-d-c-t],
there are also a couple of unique tasks found in the `.concourse/tasks` directory.

Use of the ci requires access to the following secrets:
- a [`cf-deployment-concourse-tasks`][cf-d-c-t]-style env repo (ours is [tycho-env][tycho-env])
- credentials for the release blobstore and deployment keys for the release repo.
  We store these secrets in the [`syslog-ci-private`][syslog-ci-private] repo.

You will also need a concourse to run the pipeline on.
Ours is deployed on GCP;
the deployment manifest and bbl-state can be found in [`leela-env`][leela-env].

Most people don't have access to these private repos.
The Cloud Foundry Foundation admin team can grant access.

Admin rights to the GCP projects associated with the above env-repos
is governed by membership in the cf-syslog@pivotal.io Google Group.
There are other resources (such as [durandal-env][durandal-env],
our GCP project for experimental integrations and manual testing)
associated with this, as well.

[cf-d]: https://github.com/cloudfoundry/cf-deployment
[cf-d-c-t]: https://github.com/cloudfoundry/cf-deployment-concourse-tasks
[durandal-env]: https://github.com/cloudfoundry/durandal-env
[ent-nums]: https://tools.ietf.org/html/rfc5424#section-7.2.2
[forwarder-spec-page]: https://bosh.io/jobs/syslog_forwarder?source=github.com/cloudfoundry/syslog-release
[go-installation]: https://golang.org/doc/install
[leela-env]: https://github.com/cloudfoundry/leela-env
[RFC]: https://tools.ietf.org/html/rfc5424
[sd-id]: https://tools.ietf.org/html/rfc5424#section-6.3.2
[storer-spec-page]: https://bosh.io/jobs/syslog_storer?source=github.com/cloudfoundry/syslog-release
[syslog-ci-private]: https://github.com/cloudfoundry/syslog-ci-private
[syslog-addon-ops]: https://github.com/cloudfoundry/cf-deployment/tree/master/operations/addons
[syslog-bosh-io]: https://bosh.io/releases/github.com/cloudfoundry/syslog-release
[test-readme]: tests/README.md
[tracker]: https://www.pivotaltracker.com/n/projects/2126318
[tycho-env]: https://www.github.com/cloudfoundry/tycho-env
[windows-syslog]: https://github.com/cloudfoundry/windows-syslog-release
