#!/usr/bin/env bats

setup() {
  source ./privdrop_utils.sh
}

@test "run_as_syslog runs the given command as the syslog user" {
  run run_as_syslog id -un
  [ "$status" -eq 0 ]
  [ "$output" = "syslog" ]
}

@test "run_as_syslog runs the given command as the vcap group" {
  run run_as_syslog id -gn
  [ "$status" -eq 0 ]
  [ "$output" = "vcap" ]
}

@test "run_as_syslog preserves multiple arguments without re-quoting" {
  run run_as_syslog echo one two three
  [ "$status" -eq 0 ]
  [ "$output" = "one two three" ]
}

@test "run_as_syslog preserves exported environment variables" {
  export PRIVDROP_TEST_VAR="some_value"
  run run_as_syslog sh -c 'echo "$PRIVDROP_TEST_VAR"'
  [ "$status" -eq 0 ]
  [ "$output" = "some_value" ]
}

@test "run_as_vcap runs the given command as the vcap user" {
  run run_as_vcap id -un
  [ "$status" -eq 0 ]
  [ "$output" = "vcap" ]
}

@test "run_as_vcap runs the given command as the vcap group" {
  run run_as_vcap id -gn
  [ "$status" -eq 0 ]
  [ "$output" = "vcap" ]
}

@test "run_as_vcap preserves multiple arguments without re-quoting" {
  run run_as_vcap echo one two three
  [ "$status" -eq 0 ]
  [ "$output" = "one two three" ]
}

@test "run_as_vcap preserves exported environment variables" {
  export PRIVDROP_TEST_VAR="some_value"
  run run_as_vcap sh -c 'echo "$PRIVDROP_TEST_VAR"'
  [ "$status" -eq 0 ]
  [ "$output" = "some_value" ]
}
