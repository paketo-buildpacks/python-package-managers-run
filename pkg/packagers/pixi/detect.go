// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixiinstall

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"

	"github.com/paketo-buildpacks/python-package-managers-run/pkg/build"
)

// Detect returns a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection passes when there is an environment.yml or package-list.txt file
// in the app directory, and will contribute a Build Plan that provides
// pixi-environment and requires pixi.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		projectFile, err := fs.Exists(filepath.Join(context.WorkingDir, ProjectFilename))
		if err != nil {
			return packit.DetectResult{}, packit.Fail.WithMessage("failed trying to stat %s: %w", ProjectFilename, err)
		}
		lockFile, err := fs.Exists(filepath.Join(context.WorkingDir, LockfileName))
		if err != nil {
			return packit.DetectResult{}, packit.Fail.WithMessage("failed trying to stat %s: %w", LockfileName, err)
		}

		if !projectFile && !lockFile {
			return packit.DetectResult{}, packit.Fail.WithMessage("no '%s' and '%s' found", ProjectFilename, LockfileName)
		}

		requires := []packit.BuildPlanRequirement{
			{
				Name: PixiPlanEntry,
				Metadata: build.BuildPlanMetadata{
					Build: true,
				},
			},
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: PixiEnvPlanEntry},
				},
				Requires: requires,
			},
		}, nil
	}
}
