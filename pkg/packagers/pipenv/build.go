// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipenvinstall

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

//go:generate faux --interface InstallProcess --output fakes/install_process.go
//go:generate faux --interface SitePackagesProcess --output fakes/site_packages_process.go
//go:generate faux --interface VenvDirLocator --output fakes/venv_dir_locator.go
//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

// SitePackagesProcess defines the interface for determining the site-packages path.
type SitePackagesProcess interface {
	Execute(layerPath string) (sitePackagesPath string, err error)
}

// InstallProcess defines the interface for installing the pipenv dependencies.
type InstallProcess interface {
	Execute(workingDir string, targetLayer, cacheLayer packit.Layer) error
}

// VenvDirLocator defines the interface for locating the virtual environment
// directory under a given path
type VenvDirLocator interface {
	LocateVenvDir(path string) (venvDir string, err error)
}

type PipEnvBuildParameters struct {
	InstallProcess InstallProcess
	SiteProcess    SitePackagesProcess
	VenvDirLocator VenvDirLocator
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will install the pipenv dependencies by using the Pipfile to a
// packages layer. It also makes use of a cache layer to reuse the pipenv
// cache.
func Build(
	buildParameters PipEnvBuildParameters,
	parameters pythonpackagers.CommonBuildParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		installProcess := buildParameters.InstallProcess
		siteProcess := buildParameters.SiteProcess
		venvDirLocator := buildParameters.VenvDirLocator

		sbomGenerator := parameters.SbomGenerator
		clock := parameters.Clock
		logger := parameters.Logger

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		packagesLayer, err := context.Layers.Get(PackagesLayerName)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cacheLayer, err := context.Layers.Get(CacheLayerName)
		if err != nil {
			return packit.BuildResult{}, err
		}

		planner := draft.NewPlanner()
		packagesLayer.Launch, packagesLayer.Build = planner.MergeLayerTypes(SitePackages, context.Plan.Entries)
		packagesLayer.Cache = packagesLayer.Launch || packagesLayer.Build
		cacheLayer.Cache = true

		logger.Process("Executing build process")
		duration, err := clock.Measure(func() error {
			return installProcess.Execute(context.WorkingDir, packagesLayer, cacheLayer)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		venvDir, err := venvDirLocator.LocateVenvDir(packagesLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		sitePackagesPath, err := siteProcess.Execute(packagesLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.GeneratingSBOM(packagesLayer.Path)

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

		packagesLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		packagesLayer.SharedEnv.Prepend("PATH", filepath.Join(venvDir, "bin"), ":")
		packagesLayer.SharedEnv.Prepend("PYTHONPATH", sitePackagesPath, string(os.PathListSeparator))

		logger.EnvironmentVariables(packagesLayer)

		layers := []packit.Layer{packagesLayer}
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
