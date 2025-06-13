<!--
// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>

SPDX-License-Identifier: Apache-2.0
-->

# Sub package for pipenv installation

Original implementation from `paketo-buildpacks/pipenv-install`

This sub package installs packages using pipenv and makes it available to the
application.

## Behavior

This sub package participates if `Pipfile` exists at the root of the app.

The buildpack will do the following:
- Installs the application packages to a layer made available to the app.
- Prepends the layer site-packages onto `PYTHONPATH`.
- Prepends the layer's `bin` directory to the `PATH`.

This buildpack speeds up the build process by reusing (the layer of) installed
packages from a previous build if it exists, and later cleaning up any unused
packages. For apps that do not have a `Pipfile.lock`, clean-up is not performed
to avoid the overhead of generating a lock file. Users of such apps should
either include a lock file with their app, or clear their build cache during a
rebuild to avoid any unused packages in the built image.

## Integration

The sub package provides `site-packages` as a dependency. Downstream
buildpacks can require the `site-packages` dependency by generating a [Build Plan TOML]
(https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the dependency provided by the Pipenv Install Buildpack is
  # "site-packages". This value is considered part of the public API for the
  # buildpack and will not change without a plan for deprecation.
  name = "site-packages"

  # The Pipenv Install buildpack supports some non-required metadata options.
  [requires.metadata]

    # Set the build flag to true to make the site-packages dependency available on the $PYTHONPATH/$PATH
    # for subsequent buildpacks during their build phase.
    build = true

    # Set the launch flag to true to make the site-packages dependency available on the $PYTHONPATH/$PATH
    # for the running application.
    launch = true
```

## SBOM

This buildpack can generate a Software Bill of Materials (SBOM) for the dependencies of an application.

However, this feature only works if the application already has a `Pipfile.lock` file.
This is due to a limitation in the upstream SBOM generation library (Syft).

Applications that declare their dependencies via a `Pipfile` but do not include
a `Pipfile.lock` will result in an empty SBOM. Check out the [Paketo SBOM
documentation](https://paketo.io/docs/howto/sbom/) for more information about
how to access the SBOM.
