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

# Audited events
These events are logged to the syslog drain when peformed on Bosh VM.

<table>
<thead>
  <tr>
    <th>Event</th>
    <th>Source command</th>
    <th>Log Artifact</th>
    <th>Example</th>
  </tr>
</thead>
<tbody>
  <tr>
    <td>SSH on to Bosh VM</td>
    <td>bosh ssh</td>
    <td>CEF:0|CloudFoundry|BOSH|1|agent_api|ssh</td>
    <td><pre><7>1 2018-02-28T18:36:26.488124+00:00 10.0.16.22 vcap.agent 39 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="6040e9df-afc3-4520-a6e1-0db97347951b"] 2018/02/28 18:36:26 CEF:0|CloudFoundry|BOSH|1|agent_api|ssh|1|duser=director.f762950d-2a20-4283-be8b-04e11ee11768.150cc352-9748-4b23-bcd1-7f2628ae6e82.f893ad11-8a13-4c36-8dce-861b8a1a5e7d src=10.254.50.4 spt=4222 shost=150cc352-9748-4b23-bcd1-7f2628ae6e82</pre></td>
  </tr>
  <tr>
    <td>Fail to ssh</td>
    <td>ssh</td>
    <td>sshd</td>
    <td><pre><38>1 2018-03-01T20:56:48.219606+00:00 10.0.16.22 sshd 21239 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  Connection closed by 10.0.16.25 [preauth]
<5>1 2018-03-01T20:56:48.223517+00:00 10.0.16.22 kernel - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"] [165047.702399] audit: type=1109 audit(1519937808.216:6287): pid=21239 uid=0 auid=4294967295 ses=4294967295 msg='op=PAM:bad_ident acct="?" exe="/usr/sbin/sshd" hostname=10.0.16.25 addr=10.0.16.25 terminal=ssh res=failed'</pre></td>
  </tr>
  <tr>
    <td>Finish ssh session</td>
    <td>ssh</td>
    <td>sshd:session</td>
    <td><pre><38>1 2018-03-01T21:02:37.839333+00:00 10.0.16.22 sshd 21306 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  Received disconnect from 10.0.0.5: 11: disconnected by user
<86>1 2018-03-01T21:02:37.840082+00:00 10.0.16.22 sshd 21294 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  pam_unix(sshd:session): session closed for user bosh_2c72a6a98a3441c</pre></td>
  </tr>
  <tr>
    <td>Add a User</td>
    <td>useradd</td>
    <td>useradd</td>
    <td><pre><86>1 2018-02-28T21:12:45.210868+00:00 10.0.16.22 useradd 14934 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  new group: name=myuser123, GID=1009
<86>1 2018-02-28T21:12:45.211092+00:00 10.0.16.22 useradd 14934 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  new user: name=myuser123, UID=1006, GID=1009, home=/home/myuser123, shell=
<14>1 2018-02-28T21:12:45.214741+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=ADD_GROUP msg=audit(1519852365.205:3387): pid=14934 uid=0 auid=1003 ses=8 msg='op=adding group acct="myuser123" exe="/usr/sbin/useradd" hostname=? addr=? terminal=pts/0 res=success'
</pre></td>
  </tr>
    <tr>
    <td>Delete a User</td>
    <td>userdel</td>
    <td>userdel</td>
    <td><pre><86>1 2018-02-28T21:23:04.087392+00:00 10.0.16.22 userdel 15078 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  delete user 'testin123'
<86>1 2018-02-28T21:23:04.087667+00:00 10.0.16.22 userdel 15078 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  removed group 'testin123' owned by 'testin123'
<86>1 2018-02-28T21:23:04.087810+00:00 10.0.16.22 userdel 15078 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  removed shadow group 'testin123' owned by 'testin123'
<14>1 2018-02-28T21:23:04.092054+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=DEL_GROUP msg=audit(1519852984.085:3530): pid=15078 uid=0 auid=1003 ses=8 msg='op=deleting group acct="testin123" exe="/usr/sbin/userdel" hostname=? addr=? terminal=pts/0 res=success'
<14>1 2018-02-28T21:23:04.092059+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=DEL_GROUP msg=audit(1519852984.085:3531): pid=15078 uid=0 auid=1003 ses=8 msg='op=deleting shadow group acct="testin123" exe="/usr/sbin/userdel" hostname=? addr=? terminal=pts/0 res=success'
</pre></td>
  </tr>
  <tr>
    <td>Modify a user's groups</td>
    <td>usermod -g <group> <username></td>
    <td>usermod</td>
    <td><pre><14>1 2018-02-28T23:01:32.617281+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=SYSCALL msg=audit(1519858892.612:3922): arch=c000003e syscall=82 success=yes exit=0 a0=7ffdea674220 a1=619da0 a2=7ffdea674190 a3=7ffdea673e30 items=5 ppid=13560 pid=15525 auid=4294967295 uid=0 gid=0 euid=0 suid=0 fsuid=0 egid=0 sgid=0 fsgid=0 tty=(none) ses=4294967295 comm="usermod" exe="/usr/sbin/usermod" key="identity"
<86>1 2018-02-28T22:55:12.405859+00:00 10.0.16.22 usermod 15439 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  add 'bosh_7890a75e1ecc4a6' to group 'admin'</pre></td>
  </tr>
  <tr>
    <td>User enters super user mode</td>
    <td>sudo</td>
    <td>sudo</td>
    <td><pre><14>1 2018-02-28T21:14:05.019783+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=SYSCALL msg=audit(1519852445.013:3422): arch=c000003e syscall=59 success=yes exit=0 a0=ac64a8 a1=ab4348 a2=ab8e08 a3=7fff8b51b6a0 items=2 ppid=14307 pid=14952 auid=1003 uid=0 gid=0 euid=0 suid=0 fsuid=0 egid=0 sgid=0 fsgid=0 tty=pts0 ses=8 comm="sudo" exe="/usr/bin/sudo" key="privileged"
<14>1 2018-02-28T21:14:05.019810+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=EXECVE msg=audit(1519852445.013:3422): argc=2 a0="sudo" a1="su"
<14>1 2018-02-28T21:14:05.019849+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=PATH msg=audit(1519852445.013:3422): item=0 name="/usr/bin/sudo" inode=131406 dev=08:01 mode=0104755 ouid=0 ogid=0 rdev=00:00 nametype=NORMAL
<85>1 2018-02-28T21:14:05.022169+00:00 10.0.16.22 sudo - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]      root : TTY=pts/0 ; PWD=/root ; USER=root ; COMMAND=/bin/su
<86>1 2018-02-28T21:14:05.022681+00:00 10.0.16.22 sudo - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  pam_unix(sudo:session): session opened for user root by bosh_740bae4650a640a(uid=0)
<14>1 2018-02-28T21:14:05.022946+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=USER_START msg=audit(1519852445.017:3423): pid=14952 uid=0 auid=1003 ses=8 msg='op=PAM:session_open acct="root" exe="/usr/bin/sudo" hostname=? addr=? terminal=/dev/pts/0 res=success'
</pre></td>
  </tr>
  <tr>
    <td>root changes password of user</td>
    <td>passwd <username></td>
    <td>passwd</td>
    <td><pre><14>1 2018-02-28T21:32:12.639411+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=SYSCALL msg=audit(1519853532.633:3649): arch=c000003e syscall=82 success=yes exit=0 a0=7f098fcb294a a1=7f098fcb28d4 a2=0 a3=0 items=5 ppid=14954 pid=15137 auid=1003 uid=0 gid=0 euid=0 suid=0 fsuid=0 egid=0 sgid=0 fsgid=0 tty=pts0 ses=8 comm="passwd" exe="/usr/bin/passwd" key="delete"
<85>1 2018-02-28T21:32:12.641228+00:00 10.0.16.22 passwd 15137 - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  pam_unix(passwd:chauthtok): password changed for testing1234</pre></td>
  </tr>
  <tr>
    <td>Manually stopping syslog</td>
    <td>service rsyslog stop</td>
    <td>rsyslogd</td>
    <td><pre><46>1 2018-02-28T23:17:56.118218+00:00 10.0.16.22 rsyslogd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  [origin software="rsyslogd" swVersion="8.22.0" x-pid="7345" x-info="http://www.rsyslog.com"] exiting on signal 15.</pre></td>
  </tr>
  <tr>
    <td>Manually stopping auditd</td>
    <td>service auditd stop</td>
    <td>DAEMON_END</td>
    <td><pre><14>1 2018-02-28T23:34:41.846443+00:00 10.0.16.22 audispd - - [instance@47450 director="" deployment="syslog-storer" group="syslog-forwarder" az="z1" id="e13b49d8-fb2d-48de-952d-f15071135ca6"]  node=1f045518-275e-43f4-a74c-bdd28c2c97bd type=DAEMON_END msg=audit(1519860881.846:6568): auditd normal halt, sending auid=1006 pid=16141 subj=l=4.4.0-116-generic auid=1006 pid=16107 subj=unconfined  res=success res=success</pre></td>
  </tr>
</tbody>
</table>

[auditd-man]: http://manpages.ubuntu.com/manpages/trusty/man8/auditd.8.html
[stemcell-builder]: https://github.com/cloudfoundry/bosh-linux-stemcell-builder
