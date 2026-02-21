// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetryinstall

import (
	"os"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/sbom"

	"github.com/paketo-buildpacks/python-packagers/pkg/build"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
//go:generate faux --interface InstallProcess --output fakes/install_process.go
//go:generate faux --interface PythonPathLookupProcess --output fakes/python_path_process.go

// EntryResolver defines the interface for picking the most relevant entry from
// the Buildpack Plan entries.
type EntryResolver interface {
	MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) (launch, build bool)
}

// InstallProcess defines the interface for installing the poetry dependencies.
// It returns the location of the virtual env directory.
type InstallProcess interface {
	Execute(workingDir, targetDir, cacheDir string) (string, error)
}

// PythonPathProcess defines the interface for finding the PYTHONPATH (AKA the site-packages directory)
type PythonPathLookupProcess interface {
	Execute(venvDir string) (string, error)
}

// PoetryEnvBuildParameters encapsulates the poetry specific parameters for the
// Build function
type PoetryEnvBuildParameters struct {
	EntryResolver           EntryResolver
	InstallProcess          InstallProcess
	PythonPathLookupProcess PythonPathLookupProcess
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will install the poetry dependencies by using the pyproject.toml file
// to a virtual environment layer.
func Build(
	buildParameters PoetryEnvBuildParameters,
	parameters build.CommonBuildParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		entryResolver := buildParameters.EntryResolver
		installProcess := buildParameters.InstallProcess
		pythonPathProcess := buildParameters.PythonPathLookupProcess

		sbomGenerator := parameters.SbomGenerator
		clock := parameters.Clock
		logger := parameters.Logger

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		venvLayer, err := context.Layers.Get(VenvLayerName)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cacheLayer, err := context.Layers.Get(CacheLayerName)
		if err != nil {
			return packit.BuildResult{}, err
		}

		var venvDir string
		logger.Process("Executing build process")
		duration, err := clock.Measure(func() error {
			venvDir, err = installProcess.Execute(context.WorkingDir, venvLayer.Path, cacheLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		pythonPathDir, err := pythonPathProcess.Execute(venvDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		venvLayer.Launch, venvLayer.Build = entryResolver.MergeLayerTypes(PoetryVenv, context.Plan.Entries)
		venvLayer.Cache = venvLayer.Launch || venvLayer.Build
		cacheLayer.Cache = true

		logger.GeneratingSBOM(venvLayer.Path)

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

		venvLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		venvLayer.SharedEnv.Default("POETRY_VIRTUALENVS_PATH", venvLayer.Path)
		venvLayer.SharedEnv.Prepend("PYTHONPATH", pythonPathDir, string(os.PathListSeparator))
		venvLayer.SharedEnv.Prepend("PATH", filepath.Join(venvDir, "bin"), string(os.PathListSeparator))

		logger.EnvironmentVariables(venvLayer)

		layers := []packit.Layer{venvLayer}
		if _, err := os.Stat(cacheLayer.Path); err == nil {
			if !fs.IsEmptyDir(cacheLayer.Path) {
				layers = append(layers, cacheLayer)
			}
		}

		result := packit.BuildResult{
			Layers: layers,
		}

		return result, nil
	}
}
