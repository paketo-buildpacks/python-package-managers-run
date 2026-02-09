// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"

	common "github.com/paketo-buildpacks/python-packagers/pkg/packagers/common"
)

// Detect returns a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection passes when there is an uv.lock file in the app directory,
// and will contribute a Build Plan that provides uv-environment and requires uv.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {

		lockfilePath := filepath.Join(context.WorkingDir, LockfileName)
		lockFile, err := fs.Exists(lockfilePath)
		if err != nil {
			return packit.DetectResult{}, packit.Fail.WithMessage("failed trying to stat %s: %w", LockfileName, err)
		}

		if !lockFile {
			return packit.DetectResult{}, packit.Fail.WithMessage("no 'uv.lock' found")
		}

		vendor, err := fs.Exists(filepath.Join(context.WorkingDir, "vendor"))
		if err != nil {
			return packit.DetectResult{}, packit.Fail.WithMessage("failed trying to stat %s: %w", LockfileName, err)
		}

		requires := []packit.BuildPlanRequirement{
			{
				Name: UvPlanEntry,
				Metadata: common.BuildPlanMetadata{
					Build: true,
				},
			},
		}

		if vendor {
			parser := NewLockfileParser()
			version, _ := parser.ParsePythonVersion(lockfilePath)

			requires = append(requires,
				packit.BuildPlanRequirement{
					Name: "cpython",
					Metadata: common.BuildPlanMetadata{
						Build:   true,
						Launch:  true,
						Version: version,
					},
				},
			)
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: UvEnvPlanEntry},
				},
				Requires: requires,
			},
		}, nil
	}
}
