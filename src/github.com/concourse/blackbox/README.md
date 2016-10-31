# blackbox

*forward files on to syslog*

![Black Box Flight Recorder](http://i.imgur.com/sCSNdzU.jpg)

## about

Applications often provide only a limited ability to log to syslog and often
don't log in a consistent format. I also think that syslog is an operational
concern and the application should not know about where it is logging. Blackbox
is an experiment to decouple syslogging from an application without messing
about with syslog configuration (which is tricky on BOSH VMs).

Blackbox will tail specified files and forward any new lines to a syslog
server.

## usage

```
blackbox -config config.yml
```

The configuration file schema is as follows:

``` yaml
hostname: this-host

destination:
  transport: udp
  address: logs.example.com:1234

sources:
  - path: hello.txt
    tag: hello
```

Each file can be sent with a different tag. Currently the priority and facility
are hardcoded to `INFO` and `user`. However, allowing customisation of these
per source would not be difficult.

## installation

```
go get -u github.com/concourse/blackbox/cmd/blackbox
```
