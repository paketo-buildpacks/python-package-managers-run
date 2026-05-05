// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixiinstall

const (
	// PixiEnvLayer is the name of the layer into which pixi environment is installed.
	PixiEnvLayer = "pixi-env"

	// PixiEnvCache is the name of the layer that is used as the pixi package directory.
	PixiEnvCache = "pixi-env-cache"

	// PixiEnvPlanEntry is the name of the Build Plan requirement that this buildpack provides.
	PixiEnvPlanEntry = "pixi-environment"

	// PixiPlanEntry is the name of the Build Plan requirement for the minipixi
	// dependency that this buildpack requires.
	PixiPlanEntry = "pixi"

	// LockfileShaName is the key in the Layer Content Metadata used to determine if layer
	// can be reused.
	LockfileShaName = "lockfile-sha"

	// LockfileName is the name of the export file from which the buildpack reinstalls packages
	// See https://docs.pixi.io/projects/pixi/en/latest/commands/list.html
	LockfileName = "pixi.lock"

	// ProjectFilename is the name of the pixi environment file.
	ProjectFilename = "pixi.toml"

	// PixiDefaultEnvironmentName is the name of environment created out of the project
	PixiDefaultEnvironmentName = "default"

	// PixiEnvironmentEnvVarName is the name of the environment variable used to set the
	// pixi environment to deploy
	PixiEnvironmentEnvVarName = "BP_PIXI_ENVIRONMENT_NAME"
)
