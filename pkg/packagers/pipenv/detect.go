// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipenvinstall

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"

	"github.com/paketo-buildpacks/python-packagers/pkg/build"
)

//go:generate faux --interface Parser --output fakes/parser.go

// Parser will parse python version out of Pipfile.lock.
type Parser interface {
	ParseVersion(path string) (version string, err error)
}

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection will contribute a Build Plan that provides site-packages,
// and requires cpython and pipenv at build.
func Detect(pipfileParser, pipfileLockParser Parser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		exists, err := fs.Exists(filepath.Join(context.WorkingDir, "Pipfile"))
		if err != nil {

			return packit.DetectResult{}, err
		}

		if !exists {
			return packit.DetectResult{}, packit.Fail.WithMessage("no 'Pipfile' found")
		}

		cpythonRequirement := packit.BuildPlanRequirement{
			Name: CPython,
			Metadata: build.BuildPlanMetadata{
				Build: true,
			},
		}

		lockFileExists, err := fs.Exists(filepath.Join(context.WorkingDir, "Pipfile.lock"))
		if err != nil {
			return packit.DetectResult{}, packit.Fail.WithMessage("failed trying to stat Pipfile.lock: %w", err)
		}

		if lockFileExists {
			cpythonVersion, err := pipfileLockParser.ParseVersion(context.WorkingDir)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return packit.DetectResult{}, err
				}
			}

			if cpythonVersion != "" {
				cpythonRequirement.Metadata = build.BuildPlanMetadata{
					Build:         true,
					Version:       cpythonVersion,
					VersionSource: "Pipfile.lock",
				}
			}
		} else {
			cpythonVersion, err := pipfileParser.ParseVersion(context.WorkingDir)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return packit.DetectResult{}, err
				}
			}

			if cpythonVersion != "" {
				cpythonRequirement.Metadata = build.BuildPlanMetadata{
					Build:         true,
					Version:       cpythonVersion,
					VersionSource: "Pipfile",
				}
			}
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
					cpythonRequirement,
					{
						Name: Pipenv,
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
