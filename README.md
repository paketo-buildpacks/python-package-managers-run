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

- [conda](pkg/packagers/conda/README.md) -> `environment.yml`
- [pip](pkg/packagers/pip/README.md) -> `requirements.txt`
- [pipenv](pkg/packagers/pipenv/README.md) -> `Pipfile`
- [pixi](pkg/packagers/pixi/README.md) -> `pixi.lock`
- [poetry](pkg/packagers/poetry/README.md) -> `pyproject.toml`
- [uv](pkg/packagers/uv/README.md) -> `uv.lock`

is present in the root folder.

The buildpack will do the following:

* At build time:
  - Installs the application packages to a layer made available to the app.
* At run time:
  - Does nothing

## Usage

To package this buildpack for consumption:

```bash
./scripts/package.sh --version x.x.x
```

This will create a `buildpackage.cnb` file under the build directory which you
can use to build your app as follows: `pack build <app-name> -p <path-to-app>
-b <cpython buildpack> -b <pip buildpack> -b build/buildpackage.cnb -b
<other-buildpacks..>`.

To run the unit and integration tests for this buildpack:

```bash
./scripts/unit.sh && ./scripts/integration.sh
```

To speed up local integration tests run you could clone [cpython](https://github.com/paketo-buildpacks/cpython), 
remove python versions you don't need in buildpack.toml and use the local path to the cpython folder in [`integration.json`](integration.json). Same thing applies to [python-package-managers-install](https://github.com/paketo-buildpacks/python-package-managers-install).
