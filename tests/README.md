# Usage

## Development
The following commands assume you are operating
from the top level of the repo, have go inistalled,
and have initialized and updated the blackbox submodule
as described in the main [README](../README.md).

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
scripts/test -nodes=10
```
Any arguments passed to `scripts/test`
will be passed on to Ginkgo;
here, we're running with fewer nodes than the script calls for,
to respect the limitations of our bosh-lite.
Generally, try and pick a number of nodes that evenly divides
into the number of tests you wish to run.

To run only a specific test,
see https://onsi.github.io/ginkgo/#focused-specs.

## Notes
Because this release is almost entirely composed of bosh templates,
the acceptance tests do a bosh deployment for each test.
There are helpers that make doing this easy.

If you are trying to write tests for this release,
please feel free to contact the team for assistance;
our contact info is at the top of the main [README](../README.md).
