// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	pythonpackagers "github.com/paketo-buildpacks/python-packagers"
	pkgcommon "github.com/paketo-buildpacks/python-packagers/pkg/packagers/common"
	conda "github.com/paketo-buildpacks/python-packagers/pkg/packagers/conda"
	pipinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/pip"
	pipenvinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/pipenv"
	poetryinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/poetry"
	uvinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/uv"
)

func main() {
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))

	buildParameters := pkgcommon.CommonBuildParameters{
		SbomGenerator: pkgcommon.Generator{},
		Clock:         chronos.DefaultClock,
		Logger:        logger,
	}

	packagerParameters := map[string]pythonpackagers.PackagerParameters{
		conda.CondaEnvPlanEntry: conda.CondaBuildParameters{
			Runner: conda.NewCondaRunner(pexec.NewExecutable("conda"), fs.NewChecksumCalculator(), logger),
		},
		pipinstall.Manager: pipinstall.PipBuildParameters{
			InstallProcess:      pipinstall.NewPipInstallProcess(pexec.NewExecutable("pip"), logger),
			SitePackagesProcess: pipinstall.NewSiteProcess(pexec.NewExecutable("python")),
		},
		pipenvinstall.Manager: pipenvinstall.PipEnvBuildParameters{
			InstallProcess: pipenvinstall.NewPipenvInstallProcess(pexec.NewExecutable("pipenv"), logger),
			SiteProcess:    pipenvinstall.NewSiteProcess(pexec.NewExecutable("python")),
			VenvDirLocator: pipenvinstall.NewVenvLocator(),
		},
		poetryinstall.PoetryVenv: poetryinstall.PoetryEnvBuildParameters{
			EntryResolver:           draft.NewPlanner(),
			InstallProcess:          poetryinstall.NewPoetryInstallProcess(pexec.NewExecutable("poetry"), logger),
			PythonPathLookupProcess: poetryinstall.NewPythonPathProcess(),
		},
		uvinstall.UvEnvPlanEntry: uvinstall.UvBuildParameters{
			Runner: uvinstall.NewUvRunner(pexec.NewExecutable("uv"), fs.NewChecksumCalculator(), logger),
		},
	}

	packit.Run(
		pythonpackagers.Detect(logger),
		pythonpackagers.Build(logger, buildParameters, packagerParameters),
	)
}
