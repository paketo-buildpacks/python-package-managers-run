// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipinstall

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"

	"github.com/paketo-buildpacks/python-package-managers-run/pkg/build"
)

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection will contribute a Build Plan that provides site-packages,
// and requires cpython and pip at build.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		requirementsFile := "requirements.txt"
		envRequirement, requirementEnvExists := os.LookupEnv("BP_PIP_REQUIREMENT")
		if requirementEnvExists {
			requirementsFile = envRequirement
		}

		missingRequirementFiles := []string{}
		allRequirementsFilesExist := true
		for _, filename := range strings.Split(requirementsFile, " ") {
			found, err := fs.Exists(filepath.Join(context.WorkingDir, filename))
			if err != nil {
				return packit.DetectResult{}, err
			}
			if !found {
				missingRequirementFiles = append(missingRequirementFiles, filename)
			}
			allRequirementsFilesExist = allRequirementsFilesExist && found
		}

		if !allRequirementsFilesExist {
			return packit.DetectResult{}, packit.Fail.WithMessage("requirements file not found at: '%s'", strings.Join(missingRequirementFiles, "', '"))
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{
						Name: SitePackages,
					},
					{
						Name: Manager,
					},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: CPython,
						Metadata: build.BuildPlanMetadata{
							Build: true,
						},
					},
					{
						Name: Pip,
						Metadata: build.BuildPlanMetadata{
							Build: true,
						},
					},
					{
						Name: Manager,
						Metadata: build.BuildPlanMetadata{
							Build: true,
						},
					},
				},
			},
		}, nil
	}
}
