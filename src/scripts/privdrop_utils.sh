#!/usr/bin/env bash

# run_as_syslog
#
# Exec-style replacement for `chpst -u syslog:vcap "$@"`.
# Omits --no-new-privs so blackbox retains cap_dac_read_search file capabilities.
function run_as_syslog() {
  exec setpriv --reuid=syslog --regid=vcap --clear-groups -- "$@"
}

# run_as_vcap
#
# Exec-style replacement for `chpst -u vcap:vcap "$@"`.
function run_as_vcap() {
  setpriv --reuid=vcap --regid=vcap --clear-groups --no-new-privs -- "$@"
}
