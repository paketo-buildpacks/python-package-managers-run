// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package integration_helpers

type Buildpack struct {
	ID   string
	Name string
}

type Dependency struct {
	ID      string
	Version string
}

type Metadata struct {
	Dependencies []Dependency
}

type BuildpackInfo struct {
	Buildpack Buildpack
	Metadata  Metadata
}

type TestSettings struct {
	Buildpacks struct {
		// Dependency buildpacks
		CPython struct {
			Online  string
			Offline string
		}
		PythonPackageManagersInstall struct {
			Online  string
			Offline string
		}
		BuildPlan struct {
			Online string
		}
		// This buildpack
		PythonPackageManagersRun struct {
			Online  string
			Offline string
		}
	}

	Config struct {
		CPython                      string `json:"cpython"`
		PythonPackageManagersInstall string `json:"python-package-managers-install"`
		BuildPlan                    string `json:"build-plan"`
	}
}
