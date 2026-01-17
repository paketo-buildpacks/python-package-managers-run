// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall

const (
	// UvEnvLayer is the name of the layer into which uv environment is installed.
	UvEnvLayer = "uv-env"

	// UvEnvCache is the name of the layer that is used as the uv package directory.
	UvEnvCache = "uv-env-cache"

	// UvEnvPlanEntry is the name of the Build Plan requirement that this buildpack provides.
	UvEnvPlanEntry = "uv-environment"

	// UvPlanEntry is the name of the Build Plan requirement for the uv
	// dependency that this buildpack requires.
	UvPlanEntry = "uv"

	// LockfileShaName is the key in the Layer Content Metadata used to determine if layer
	// can be reused.
	LockfileShaName = "lockfile-sha"

	// LockfileName is the name of the export file from which the buildpack reinstalls packages
	LockfileName = "uv.lock"
)
