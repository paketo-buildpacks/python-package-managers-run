// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"

	pip "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pip"
	poetry "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/poetry"
	uv "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/uv"
)

const (
	Hatchling  = "hatchling.build"
	Setuptools = "setuptools.build_meta"
	Flit       = "flit_core.buildapi"
	PDG        = "pdm.backend"
	UvBuild    = "uv_build"
	Poetry     = "poetry.core.masonry.api"

	ProjectFile = "pyproject.toml"
)

type BuildSystem struct {
	Requires     []string `toml:"requires"`
	BuildBackend string   `toml:"build-backend"`
}

type Project struct {
	BuildSystem BuildSystem `toml:"build-system"`
}

type PyProjectHandler struct {
}

var InstallerMap = map[string]string{
	Setuptools: pip.Pip,
	Poetry:     poetry.Poetry,
	UvBuild:    uv.UvPlanEntry,
	"":         poetry.Poetry, // To keep compatibility with original implementation and poetry v1
}

func NewPyProjectHandler() PyProjectHandler {
	return PyProjectHandler{}
}

func (p *PyProjectHandler) GetBuildBackend(projectPath string) (string, error) {
	var project Project
	_, err := toml.DecodeFile(filepath.Join(projectPath, ProjectFile), &project)
	if err != nil {
		return "", err
	}

	return project.BuildSystem.BuildBackend, nil
}

func (p *PyProjectHandler) GetInstaller(projectPath string) (string, error) {
	backend, err := p.GetBuildBackend(projectPath)
	if err != nil {
		return "", err
	}

	if backend == "" {
		found, err := fs.Exists(filepath.Join(projectPath, uv.LockfileName))
		if err != nil {
			return "", err
		}

		if found {
			backend = UvBuild
		}

	}

	installer, ok := InstallerMap[backend]
	if !ok {
		return "", fmt.Errorf("unsupported backend: %s", backend)
	}

	return installer, nil
}

func (p *PyProjectHandler) Detect(installer string, context packit.DetectContext) (packit.DetectResult, error) {
	var result packit.DetectResult
	var err error

	switch installer {
	case pip.Pip:
		result, err = pip.Detect()(context)

	case poetry.Poetry:
		result, err = poetry.Detect()(context)

	case uv.UvPlanEntry:
		result, err = uv.Detect()(context)

	default:
		result = packit.DetectResult{}
		err = fmt.Errorf("unsupported installer: %s", installer)
	}

	return result, err
}
