<!--
SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>

SPDX-License-Identifier: Apache-2.0
-->

# Sub package for pixi environment installation

This sub package runs commands to install a pixi environment. It installs the
pixi environment into a layer which makes it available for subsequent
buildpacks and in the final running container.

## Behavior

This sub package participates when there is an `pixi.lock` or
`pixi.toml` file in the app directory.

The buildpack will do the following:

* At build time:
    - Requires that pixi has already been installed in the build container
    - Install the pixi environment and stores it in a layer
    - Reuses the cached pixi environment layer from a previous build if neither
      the project nor lock file has changed.
* At run time:
    - Does nothing

## Integration

This sub package provides `pixi-environment` as a dependency. Downstream
buildpacks can require the `pixi-environment` dependency by generating a
[Build Plan TOML]
(https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]
# The name of the Pixi Install dependency is "pixi-environment". This value is
# considered part of the public API for the buildpack and will not change
# without a plan for deprecation.
name = "pixi-environment"

# The Pixi Install buildpack supports some non-required metadata options.
[requires.metadata]

# Setting the build flag to true will ensure that the pixi environment
# layer is available for subsequent buildpacks during their build phase.
# If you are writing a buildpack that needs the pixi environment
# during its build process, this flag should be set to true.
build = true

# Setting the launch flag to true will ensure that the pixi environment is
# available to the running application. If you are writing an application
# that needs to use the pixi environment at runtime, this flag should be set to true.
launch = true
```
