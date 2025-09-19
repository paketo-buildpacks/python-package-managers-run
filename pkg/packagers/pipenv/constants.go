// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipenvinstall

// SitePackages is the name of the dependency provided by the Pipenv Install
// buildpack.
const SitePackages = "site-packages"

// Manager is the name of a provided/self-consumed entry plan to differentiate
// with pipenv as no Metadata are available to provides entries
const Manager = "manager-pipenv"

// CPython is the name of the python runtime dependency provided by the CPython
// buildpack: https://github.com/paketo-buildpacks/cpython.
const CPython = "cpython"

// Pipenv is the name of the dependency provided by the Pipenv buildpack:
// https://github.com/paketo-buildpacks/pipenv.
const Pipenv = "pipenv"

// The layer name for packages layer. This layer is where dependencies are
// installed to.
const PackagesLayerName = "packages"

// The layer name for cache layer. This layer holds the pipenv cache.
const CacheLayerName = "cache"
