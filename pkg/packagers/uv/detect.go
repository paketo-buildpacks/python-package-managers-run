// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

// Detect returns a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection passes when there is an environment.yml or package-list.txt file
// in the app directory, and will contribute a Build Plan that provides
// conda-environment and requires conda.
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
				Metadata: map[string]interface{}{
					"build": true,
				},
			},
		}

		if vendor {
			parser := NewLockfileParser()
			version, _ := parser.ParsePythonVersion(lockfilePath)

			requires = append(requires,
				packit.BuildPlanRequirement{
					Name: "cpython",
					Metadata: map[string]interface{}{
						"build":   true,
						"launch":  true,
						"version": version,
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
