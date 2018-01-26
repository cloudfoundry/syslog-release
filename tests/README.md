# Usage

## Development
The following commands assume you are operating
from the top level of the repo.

First, you'll need to setup a bosh-lite and login to it.
If you don't have a bosh-lite running
and aliased as `vbox` already:
```sh
scripts/setup-bosh-lite-for-tests.sh
```

If you don't already have BOSH credential
environment variables in your session:
```
scripts/export-bosh-lite-creds.sh
```

To then run the tests locally:
```sh
BOSH_ENVIRONMENT=vbox scripts/test
```
