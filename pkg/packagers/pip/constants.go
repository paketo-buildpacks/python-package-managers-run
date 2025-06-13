// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipinstall

// SitePackages is the name of the dependency provided by the Pip Install
// buildpack.
const SitePackages = "site-packages"

// Manager is the name of a provided/self-consumed entry plan to differentiate
// with pipenv as no Metadata are available to provides entries
const Manager = "manager-pip"

// CPython is the name of the python runtime dependency provided by the CPython
// buildpack: https://github.com/paketo-buildpacks/cpython.
const CPython = "cpython"

// Pip is the name of the dependency provided by the Pip buildpack:
// https://github.com/paketo-buildpacks/pip.
const Pip = "pip"

// The layer name for packages layer. This layer is where dependencies are
// installed to.
const PackagesLayerName = "packages"

// The layer name for cache layer. This layer holds the pip cache.
const CacheLayerName = "cache"
