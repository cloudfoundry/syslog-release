A simple test runner. In case of failure, deployment resources are not deleted.

    ./tests/execute

    # or for just one suite of tests
    ./tests/execute defaults

The [`manifest.yml`](manifest.yml) is the base deployment manifest with a single storer and single forwarder. Individual test suites must have a `manifest-ops.yml` file in their directory which may further manipulate the base manifest. A fresh deployment is created for each suite, and then the `test-*` files are executed against it.


# Setup

To run, ensure you have [bosh-cli](https://bosh.io/docs/cli-v2.html) installed and configured the environment with...

 * **`BOSH_ENVIRONMENT`** - the environment for test deployments (you must already be logged in)
 * `BOSH_GW_HOST`, `BOSH_GW_USER`, `BOSH_GW_PRIVATE_KEY` - SSH gateway details (if necessary)

And that a cloud-config is configured with...

 * VM Type `default`
 * Network `default`
 * Availability Zone `z1`

And a stemcell is uploaded.


## bosh-lite

If you're using bosh-lite...

    $ bosh upload-stemcell \
      --sha1=7e8d841c5f4d736285ce21a1d582a645c2830cbf \
      https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=3363.9
    $ bosh update-cloud-config <( wget -qO https://raw.githubusercontent.com/cloudfoundry/bosh-deployment/master/warden/cloud-config.yml )
