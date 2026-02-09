// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall

import (
	"os"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/sbom"

	pythonpackagers "github.com/paketo-buildpacks/python-packagers/pkg/packagers/common"
)

//go:generate faux --interface Runner --output fakes/runner.go
//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

// Runner defines the interface for setting up the uv environment.
type Runner interface {
	Execute(uvEnvPath string, uvCachePath string, workingDir string) error
	ShouldRun(workingDir string, metadata map[string]interface{}) (bool, string, error)
}

// UvBuildParameters encapsulates the uv specific parameters for the
// Build function
type UvBuildParameters struct {
	Runner Runner
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build updates the uv environment and stores the result in a layer. It may
// reuse the environment layer from a previous build, depending on conditions
// determined by the runner.
func Build(
	buildParameters UvBuildParameters,
	parameters pythonpackagers.CommonBuildParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		runner := buildParameters.Runner

		sbomGenerator := parameters.SbomGenerator
		clock := parameters.Clock
		logger := parameters.Logger

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		uvLayer, err := context.Layers.Get(UvEnvLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		uvCacheLayer, err := context.Layers.Get(UvEnvCache)
		if err != nil {
			return packit.BuildResult{}, err
		}

		run, sha, err := runner.ShouldRun(context.WorkingDir, uvLayer.Metadata)
		if err != nil {
			return packit.BuildResult{}, err
		}

		if run {
			uvLayer, err = uvLayer.Reset()
			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.Process("Executing build process")
			duration, err := clock.Measure(func() error {
				return runner.Execute(uvLayer.Path, uvCacheLayer.Path, context.WorkingDir)
			})
			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.Action("Completed in %s", duration.Round(time.Millisecond))
			logger.Break()

			logger.GeneratingSBOM(uvLayer.Path)

			var sbomContent sbom.SBOM
			duration, err = clock.Measure(func() error {
				sbomContent, err = sbomGenerator.Generate(context.WorkingDir)
				return err
			})
			if err != nil {
				return packit.BuildResult{}, err
			}
			logger.Action("Completed in %s", duration.Round(time.Millisecond))
			logger.Break()

			logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)

			uvLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
			if err != nil {
				return packit.BuildResult{}, err
			}

			uvLayer.SharedEnv.Prepend("PATH", filepath.Join(uvLayer.Path, "venv", "bin"), string(os.PathListSeparator))

			logger.EnvironmentVariables(uvLayer)

			uvLayer.Metadata = map[string]interface{}{
				LockfileShaName: sha,
			}
		} else {
			logger.Process("Reusing cached layer %s", uvLayer.Path)
			logger.Break()
		}

		planner := draft.NewPlanner()
		uvLayer.Launch, uvLayer.Build = planner.MergeLayerTypes(UvEnvPlanEntry, context.Plan.Entries)
		uvLayer.Cache = uvLayer.Build
		uvCacheLayer.Cache = true

		layers := []packit.Layer{uvLayer}
		if _, err := os.Stat(uvCacheLayer.Path); err == nil {
			if !fs.IsEmptyDir(uvCacheLayer.Path) {
				layers = append(layers, uvCacheLayer)
			}
		}

		return packit.BuildResult{
			Layers: layers,
		}, nil
	}
}
