<!--
// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>

SPDX-License-Identifier: Apache-2.0
-->

# Sub package for uv installation

This sub package installs packages using uv and makes it available to the
application.

## Behavior
This sub package participates if `uv.lock` exists at the root the app.

The buildpack will do the following:
* At build time:
  - Creates a virtual environment based on the python version configured
  - Installs the application packages to a layer made available to the app.
  - If a vendor directory is available:
    - the cpython buildpack will be used to provide the base python environment
    - will attempt to run `uv pip install` in an offline manner.
* At run time:
  - Does nothing

## Integration

This sub package provides `uv-environment` as a dependency. Downstream buildpacks
can require the `uv-environment` dependency by generating a [Build Plan TOML]
(https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the dependency provided by this buildpack is
  # "uv-environment". This value is considered part of the public API for the
  # buildpack and will not change without a plan for deprecation.
  name = "uv-environment"

  # This buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the site-packages
    # dependency is available on the $PYTHONPATH for subsequent
    # buildpacks during their build phase. If you are writing a buildpack that
    # needs site-packages during its build process, this flag should be
    # set to true.
    build = true

    # Setting the launch flag to true will ensure that the site-packages
    # dependency is available on the $PYTHONPATH for the running
    # application. If you are writing an application that needs site-packages
    # at runtime, this flag should be set to true.
    launch = true
```

## Configuration

### `BP_UV_FIND_LINKS`

The `BP_UV_FIND_LINKS` variable allows you to specify one or more directories
to pass to `--find-links`. This should be a local path or `file://` URL.

```shell
BP_UV_FIND_LINKS=./vendor-dir
```
