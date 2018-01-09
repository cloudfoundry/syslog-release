#! /usr/bin/env bash

# NB: you will need to source this file if you want these vars in your session

export BOSH_CLIENT=admin
export BOSH_CLIENT_SECRET
BOSH_CLIENT_SECRET=$(bosh int ~/deployments/vbox/creds.yml --path /admin_password)
