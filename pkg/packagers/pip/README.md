<!--
// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>

SPDX-License-Identifier: Apache-2.0
-->

# Sub package for pip installation

Original implementation from `paketo-buildpacks/pip-install`

This sub package installs packages using pip and makes it available to the
application.

## Behavior
This sub package participates if `requirements.txt` exists at the root the app.

The buildpack will do the following:
* At build time:
  - Installs the application packages to a layer made available to the app.
  - Prepends the layer site-packages onto `PYTHONPATH`.
  - If a vendor directory is available, will attempt to run `pip install` in an offline manner.
* At run time:
  - Does nothing

## Integration

This sub package provides `site-packages` as a dependency. Downstream buildpacks
can require the `site-packages` dependency by generating a [Build Plan TOML]
(https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the dependency provided by the Pip Install Buildpack is
  # "site-packages". This value is considered part of the public API for the
  # buildpack and will not change without a plan for deprecation.
  name = "site-packages"

  # The Pip Install buildpack supports some non-required metadata options.
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

### `BP_PIP_DEST_PATH`

The `BP_PIP_DEST_PATH` variable allows you to specify a custom vendor directory.
This should be a directory underneath the working directory.
Will use `./vendor` if not provided.

```shell
BP_PIP_DEST_PATH=my/custom/vendor-dir
```

### `BP_PIP_REQUIREMENT`

The `BP_PIP_REQUIREMENT` variable allows you to specify a custom pip requirement path.
This should be a file underneath the working directory.
Will use `./requirements.txt` if not provided.

```shell
BP_PIP_REQUIREMENT=requirements-dev.txt
```

### `BP_PIP_FIND_LINKS`

The `BP_PIP_FIND_LINKS` variable allows you to specify one or more directories
to pass to `--find-links`. This should be a local path or `file://` URL.

```shell
BP_PIP_FIND_LINKS=./vendor-dir
```

### `PIP_<UPPER_LONG_NAME>`

It is worth noting that the `PIP_<UPPER_LONG_NAME>` configuration is respected
by this buildpack and can be used to tweak the build time CLI properties for
Pip as documented in [Pip's
configuration](https://pip.pypa.io/en/stable/topics/configuration/#environment-variables).
