<!--
SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>

SPDX-License-Identifier: Apache-2.0
-->

# Python Packagers Cloud Native Buildpack

The Paketo Buildpack for Python Packagers is a Cloud Native Buildpack that
installs packages using the adequate tool selected based on the content of the
application sources and makes it available to it.

The buildpack is published for consumption at
`gcr.io/paketo-buildpacks/python-package-managers-run` and
`paketobuildpacks/python-package-managers-run`.

## Behavior
This buildpack participates if one of the following detection succeeds:

- [conda](pkg/conda/README.md) -> `environment.yml`
- [pip](pkg/pip/README.md) -> `requirements.txt`
- [pipenv](pkg/pipenv/README.md) -> `Pipfile`
- [pixi](pkg/pixi/README.md) -> `pixi.lock`
- [poetry](pkg/poetry/README.md) -> `pyproject.toml`
- [uv](pkg/uv/README.md) -> `uv.lock`

is present in the root folder.

The buildpack will do the following:
* At build time:
  - Installs the application packages to a layer made available to the app.
* At run time:
  - Does nothing

## Configuration

### `BP_ENABLE_PACKAGE_MANAGERS`

The `BP_ENABLE_PACKAGE_MANAGERS` environment variable allows you to force the use
of this buildpack for all the supported package managers. It works in tandem
with `python-start`. `python-start` will add a requirement that is fulfilled by
this buildpack.

It is currently used as an opt-in to allow Paketo users to do tests before the
old buildpacks get retired.

```shell
BP_ENABLE_PACKAGE_MANAGERS=true
```

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh --version x.x.x
```
This will create a `buildpackage.cnb` file under the build directory which you
can use to build your app as follows: `pack build <app-name> -p <path-to-app>
-b <cpython buildpack> -b <pip buildpack> -b build/buildpackage.cnb -b
<other-buildpacks..>`.

To run the unit and integration tests for this buildpack:
```
$ ./scripts/unit.sh && ./scripts/integration.sh
```
