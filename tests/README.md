# Usage

## Development

The following commands assume you are operating
from the top level of the repo, have go installed,
and have initialized and updated the blackbox submodule
as described in the main [README](../README.md).

Ensure that you have a BOSH Director deployed and your local environment is
configured to point to the BOSH Director.

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

You can set `STEMCELL_OS` to run the tests with arbitrary stemcells.
Any valid value for `stemcell.os` in the BOSH manifest should work -
the tests end up interpolating the env var into the test manifests.

## Notes

Because this release is almost entirely composed of bosh templates,
the acceptance tests do a bosh deployment for each test.
There are helpers that make doing this easy.

If you are trying to write tests for this release,
please feel free to contact the team for assistance;
our contact info is at the top of the main [README](../README.md).
