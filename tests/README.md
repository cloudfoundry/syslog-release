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
```sh
scripts/export-bosh-lite-creds.sh
```

To then run the tests locally:
```sh
BOSH_ENVIRONMENT=vbox scripts/test
```

## Notes
Because this release is almost entirely composed of bosh templates,
the acceptance tests do a bosh deployment for each test.
There are helpers that make doing this easy.

If you are trying to write tests for this release,
please feel free to contact the team for assistance;
our contact info is at the top of the main [README](../README.md).
