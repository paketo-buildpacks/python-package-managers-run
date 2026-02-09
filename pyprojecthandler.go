// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/packit/v2"

	common "github.com/paketo-buildpacks/python-packagers/pkg/packagers/common"
	pip "github.com/paketo-buildpacks/python-packagers/pkg/packagers/pip"
	poetry "github.com/paketo-buildpacks/python-packagers/pkg/packagers/poetry"
	uv "github.com/paketo-buildpacks/python-packagers/pkg/packagers/uv"
)

const (
	Hatchling  = "hatchling.build"
	Setuptools = "setuptools.build_meta"
	Flit       = "flit_core.buildapi"
	PDG        = "pdm.backend"
	UvBuild    = "uv_build"
	Poetry     = "poetry.core.masonry.api"
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

func (p *PyProjectHandler) GetBuildBackend(path string) (string, error) {
	var project Project
	_, err := toml.DecodeFile(path, &project)
	if err != nil {
		return "", err
	}

	return project.BuildSystem.BuildBackend, nil
}

func (p *PyProjectHandler) GetInstaller(path string) (string, error) {
	backend, err := p.GetBuildBackend(path)
	if err != nil {
		return "", err
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
		result = packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{
						Name: pip.SitePackages,
					},
					{
						Name: pip.Manager,
					},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: pip.CPython,
						Metadata: common.BuildPlanMetadata{
							Build: true,
						},
					},
					{
						Name: pip.Pip,
						Metadata: common.BuildPlanMetadata{
							Build: true,
						},
					},
					{
						Name: pip.Manager,
						Metadata: common.BuildPlanMetadata{
							Build: true,
						},
					},
				},
			},
		}

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
