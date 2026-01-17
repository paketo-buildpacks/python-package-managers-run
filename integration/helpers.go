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
		Miniconda struct {
			Online  string
			Offline string
		}
		CPython struct {
			Online  string
			Offline string
		}
		Pip struct {
			Online  string
			Offline string
		}
		Pipenv struct {
			Online  string
			Offline string
		}
		Poetry struct {
			Online string
		}
		BuildPlan struct {
			Online string
		}
		PythonInstallers struct {
			Online  string
			Offline string
		}
		// This buildpack
		PythonPackagers struct {
			Online  string
			Offline string
		}
	}

	Config struct {
		Miniconda        string `json:"miniconda"`
		CPython          string `json:"cpython"`
		Pip              string `json:"pip"`
		Pipenv           string `json:"pipenv"`
		Poetry           string `json:"poetry"`
		PythonInstallers string `json:"python-installers"`
		BuildPlan        string `json:"build-plan"`
	}
}
