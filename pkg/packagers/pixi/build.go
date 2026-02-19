// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixiinstall

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

// Runner defines the interface for setting up the pixi environment.
type Runner interface {
	Execute(pixiEnvPath string, pixiCachePath string, workingDir string) error
	ShouldRun(workingDir string, metadata map[string]interface{}) (bool, string, error)
}

type SBOMGenerator interface {
	Generate(dir string) (sbom.SBOM, error)
}

// PixiBuildParameters encapsulates the pixi specific parameters for the
// Build function
type PixiBuildParameters struct {
	Runner Runner
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build creates the pixi environment and stores the result in a layer. It may
// reuse the environment layer from a previous build, depending on conditions
// determined by the runner.
func Build(
	buildParameters PixiBuildParameters,
	parameters pythonpackagers.CommonBuildParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		runner := buildParameters.Runner
		sbomGenerator := parameters.SbomGenerator
		clock := parameters.Clock
		logger := parameters.Logger

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		pixiLayer, err := context.Layers.Get(PixiEnvLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		pixiCacheLayer, err := context.Layers.Get(PixiEnvCache)
		if err != nil {
			return packit.BuildResult{}, err
		}

		run, sha, err := runner.ShouldRun(context.WorkingDir, pixiLayer.Metadata)
		if err != nil {
			return packit.BuildResult{}, err
		}

		if run {
			pixiLayer, err = pixiLayer.Reset()
			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.Process("Executing build process")

			duration, err := clock.Measure(func() error {
				return runner.Execute(pixiLayer.Path, pixiCacheLayer.Path, context.WorkingDir)
			})
			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.Action("Completed in %s", duration.Round(time.Millisecond))
			logger.Break()

			logger.GeneratingSBOM(pixiLayer.Path)

			var sbomContent sbom.SBOM
			duration, err = clock.Measure(func() error {
				// Syft does not support pixi yet
				sbomContent, err = sbomGenerator.Generate(pixiLayer.Path)
				return err
			})
			if err != nil {
				return packit.BuildResult{}, err
			}
			logger.Action("Completed in %s", duration.Round(time.Millisecond))
			logger.Break()

			logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)

			pixiLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
			if err != nil {
				return packit.BuildResult{}, err
			}

			pixiLayer.Metadata = map[string]interface{}{
				"lockfile-sha": sha,
			}

			pixiLayer.SharedEnv.Prepend("PATH", filepath.Join(pixiLayer.Path, PixiEnvironmentName, "bin"), string(os.PathListSeparator))

			logger.EnvironmentVariables(pixiLayer)

		} else {
			logger.Process("Reusing cached layer %s", pixiLayer.Path)
			logger.Break()
		}

		planner := draft.NewPlanner()
		pixiLayer.Launch, pixiLayer.Build = planner.MergeLayerTypes(PixiEnvPlanEntry, context.Plan.Entries)
		pixiLayer.Cache = pixiLayer.Build
		pixiCacheLayer.Cache = true

		layers := []packit.Layer{pixiLayer}
		if _, err := os.Stat(pixiCacheLayer.Path); err == nil {
			if !fs.IsEmptyDir(pixiCacheLayer.Path) {
				layers = append(layers, pixiCacheLayer)
			}
		}

		return packit.BuildResult{
			Layers: layers,
		}, nil
	}
}
