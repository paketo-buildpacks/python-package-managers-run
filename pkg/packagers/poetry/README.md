<!--
// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>

SPDX-License-Identifier: Apache-2.0
-->

# Sub package for poetry installation

Original implementation from `paketo-buildpack/poetry-install`

This sub package installs packages using [Poetry](https://python-poetry.org/) and
makes the installed packages available to the application.

## Behavior

This sub package participates if `pyproject.toml` exists at the root the app.

The buildpack will do the following:
* At build time:
  - Creates a virtual environment, installs the application packages to it,
    and makes this virtual environment available to the app via a layer called `poetry-venv`.
  - Configures `poetry` to locate this virtual environment via the
    environment variable `POETRY_VIRTUAL_ENVS_PATH`.
  - Prepends the layer `poetry-venv` onto `PYTHONPATH`.
  - Prepends the `bin` directory of the `poetry-venv` layer to the `PATH` environment variable.
* At run time:
  - Does nothing

## Integration

This sub package provides `poetry-venv` as a dependency. Downstream buildpacks
can require the `poetry-venv` dependency by generating a [Build Plan TOML]
(https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the dependency provided by the Poetry Install Buildpack is
  # "poetry-venv". This value is considered part of the public API for the
  # buildpack and will not change without a plan for deprecation.
  name = "poetry-venv"

  # The Poetry Install buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the poetry-venv
    # dependency is available on the $PYTHONPATH for subsequent
    # buildpacks during their build phase. If you are writing a buildpack that
    # needs poetry-venv during its build process, this flag should be
    # set to true.
    build = true

    # Setting the launch flag to true will ensure that the poetry-venv
    # dependency is available on the $PYTHONPATH for the running
    # application. If you are writing an application that needs poetry-venv
    # at runtime, this flag should be set to true.
    launch = true
```

## Known issues and limitations

* This buildpack will not work in an offline/air-gapped environment: vendoring
  of dependencies is not supported. This is a limitation of `poetry` - which
  itself does not support vendoring dependencies.
