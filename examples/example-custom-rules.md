# Example Custom Rules
You can use the `syslog.custom_rule` property
to insert custom RSYSLOG configuration
before the forwarding rule.
Note that such rules only affect
what is forwarded,
not what the jobs emit
or write to disk.

While this offers a very broad set of options,
typically you'll probably want
to do something simple,
like filter out certain logs,
or only forward certain logs.

This file has a few simple examples
to help you accomplish these ends.

These examples are all in [rainerscript][rainerscript-docs].
There are at least two
other configuration syntaxes supported by RSYSLOG.
They probably work, but,
our testing and examples focus on rainerscript.

Note: when specifying these rules
in the bosh manifest, you'll need to use either single quotes,
or the yaml "pipe" (`|`) syntax for multi-line strings.
Double quotes or default/detected strings are likely to be invalid yaml,
because of the characters used in the rules.

## Examples
### Dropping `DEBUG` Logs
```
if ($msg contains "DEBUG") then stop
```

### Forwarding _Only_ Logs From a Certain Job
```
if ($app-name != "uaa") then stop
```

### Writing Certain Logs to a Local File
Sometimes it is useful to write logs
to a local file.
This rule will require you to `ssh`
or otherwise access the local filesystem
in order to see the results.
If you put this rule after
some other rule, you can use it to test
the effectiveness of said other rule.

Please note that the entire config
will fail to operate
if an invalid "forwarding" rule is present.
(It's the `*.*` rule near the bottom of the config.)
Comment it out if you want to test locally
without a valid forwarding target.

```
if ($app-name != "uaa") then {
  action(
    type="omfile"
    File="/var/log/experimental.log"
  )
}
```

#Forwarding to additional remotes

If you want to forward logs to remotes in addition to the remote set in the manifest,
you can use the custom rule field to do so. To send all logs to additional remotes, 
set a remote using a forwarding rule. For example, if you wanted to send log lines
to a server at 127.0.0.1 over port 514 you could set your custom_rule property to:

```
#udp address
*.* @127.0.0.1:514;SyslogForwarderTemplate

#tcp address
*.* @@127.0.0.1:514;SyslogForwarderTemplate
```

If you wish to send a subset of messages, you can forward using conditionals, as well
as the forwarding action type. This is just a slightly different syntax for forwarding using addresses.
For example, if you wanted to forward all messages that contain the word test to a syslog server located at
127.0.0.1 over port 514 using tcp, you can use the following rule in your custom_rule property. 
```
if ($msg contains "test") then action(type="omfwd" Target="127.0.0.1" Port="514" Protocol="tcp" template="SyslogForwarderTemplate")
```

You can combine these conditionals with and: 

```
if ($msg contains "test" and $msg contains "IMPORTANT") then action(type="omfwd" Target="127.0.0.1" Port="514" Protocol="tcp" template="SyslogForwarderTemplate")
```

If you want to then not send those log lines to the primary syslog reciever, you can then issue a custom rule afterwards
to stop processing those messages. 

```
if ($msg contains "test") then action(type="omfwd" Target="127.0.0.1" Port="514" Protocol="tcp" template="SyslogForwarderTemplate")
if ($msg contains "test") then stop
```

### Configuring Global Properties
It is possible to override global rsyslog config.
This can be complicated, and may not always work as expected.

For instance, the following rule will increase the [maximum message size][global-config-doc]:
```
$MaxMessageSize 10k
```
This is a very flexible option.
It can be useful, as in the above case,
for [working around issues][blackbox-trunc-issue]
or prototyping changes to syslog-release.
We can't necessarily support everything you might do with this.
If you find yourself always configuring a certain parameter,
please consider raising an issue to get it promoted to the spec,
where we can test and document it.

## Further Notes
In most of the above examples,
the stop directive (`stop`)
is used to prevent any further processing
of a log line matching a conditional.

Many other actions and conditions are possible.

You can find config docs for RSYSLOG [here][rsyslog-config-docs].

Docs for rainerscript can be found [here][rainerscript-docs].

[rainerscript-docs]: http://www.rsyslog.com/doc/v8-stable/rainerscript/index.html
[rsyslog-config-docs]: http://www.rsyslog.com/doc/v8-stable/configuration/basic_structure.html
[global-config-doc]: https://www.rsyslog.com/doc/v8-stable/configuration/global/index.html
[blackbox-trunc-issue]: https://github.com/cloudfoundry/syslog-release/issues/37
