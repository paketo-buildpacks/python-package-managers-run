// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	conda "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/conda"
	pipinstall "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pip"
	pipenvinstall "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pipenv"
	pixiinstall "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pixi"
)

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// If this buildpack detects files that indicate your app is a Python project,
// it will pass detection.
func Detect(logger scribe.Emitter) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		logger.Title("Checking for pyproject.toml")
		pyprojectPath := filepath.Join(context.WorkingDir, "pyproject.toml")
		found, err := fs.Exists(pyprojectPath)
		if err != nil {
			return packit.DetectResult{}, err
		}
		if found {
			parser := NewPyProjectHandler()
			installer, err := parser.GetInstaller(context.WorkingDir)
			if err != nil {
				return packit.DetectResult{}, err
			}
			logger.Detail("Doing detection for: %s", installer)
			return parser.Detect(installer, context)
		}

		logger.Title("Checking for pip")
		pipResult, err := pipinstall.Detect()(context)

		if err == nil {
			return pipResult, nil
		} else {
			logger.Detail("%s", err)
		}

		logger.Title("Checking for pipenv")
		pipenvResult, err := pipenvinstall.Detect(
			pipenvinstall.NewPipfileParser(),
			pipenvinstall.NewPipfileLockParser(),
		)(context)

		if err == nil {
			return pipenvResult, nil
		} else {
			logger.Detail("%s", err)
		}

		logger.Title("Checking for pixi")
		pixiResult, err := pixiinstall.Detect()(context)

		if err == nil {
			return pixiResult, nil
		} else {
			logger.Detail("%s", err)
		}

		logger.Title("Checking for conda")
		condaResult, err := conda.Detect()(context)

		if err == nil {
			return condaResult, nil
		} else {
			logger.Detail("%s", err)
		}

		return packit.DetectResult{}, packit.Fail.WithMessage("No python packager manager related files found")
	}
}
