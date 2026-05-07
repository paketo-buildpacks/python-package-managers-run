// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers_test

import (
	"bytes"
	// "os"
	// "path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	pythonpackagers "github.com/paketo-buildpacks/python-package-managers-run"
	"github.com/paketo-buildpacks/python-package-managers-run/pkg/build"
	conda "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/conda"
	condafakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/conda/fakes"
	pipinstall "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pip"
	pipfakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pip/fakes"
	pipenvinstall "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pipenv"
	pipenvfakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pipenv/fakes"
	poetryinstall "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/poetry"
	poetryfakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/poetry/fakes"
	uvinstall "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/uv"
	uvfakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/uv/fakes"
	sbomfakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/sbom/fakes"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string

		buffer       *bytes.Buffer
		logger       scribe.Emitter
		buildFunc    packit.BuildFunc
		buildContext packit.BuildContext

		// common
		sbomGenerator *sbomfakes.SBOMGenerator

		// conda
		runner *condafakes.Runner

		// pip
		pipInstallProcess      *pipfakes.InstallProcess
		pipSitePackagesProcess *pipfakes.SitePackagesProcess

		// pipenv
		pipenvInstallProcess      *pipenvfakes.InstallProcess
		pipenvSitePackagesProcess *pipenvfakes.SitePackagesProcess
		pipenvVenvDirLocator      *pipenvfakes.VenvDirLocator

		// poetry
		poetryEntryResolver     *poetryfakes.EntryResolver
		poetryInstallProcess    *poetryfakes.InstallProcess
		poetryPythonPathProcess *poetryfakes.PythonPathLookupProcess

		// uv
		uvRunner *uvfakes.Runner

		buildParameters build.CommonBuildParameters

		plans []packit.BuildpackPlan
	)

	it.Before(func() {
		layersDir = t.TempDir()
		workingDir = t.TempDir()
		cnbDir = t.TempDir()

		buffer = bytes.NewBuffer(nil)
		logger = scribe.NewEmitter(buffer)

		sbomGenerator = &sbomfakes.SBOMGenerator{}
		sbomGenerator.GenerateCall.Returns.SBOM = sbom.SBOM{}

		// conda
		runner = &condafakes.Runner{}
		runner.ShouldRunCall.Returns.Bool = true
		runner.ShouldRunCall.Returns.String = "some-sha"

		// pip
		pipInstallProcess = &pipfakes.InstallProcess{}
		pipSitePackagesProcess = &pipfakes.SitePackagesProcess{}
		pipSitePackagesProcess.ExecuteCall.Returns.SitePackagesPath = "some-site-packages-path"

		// pipenv
		pipenvInstallProcess = &pipenvfakes.InstallProcess{}
		pipenvSitePackagesProcess = &pipenvfakes.SitePackagesProcess{}
		pipenvSitePackagesProcess.ExecuteCall.Returns.SitePackagesPath = "some-site-packages-path"
		pipenvVenvDirLocator = &pipenvfakes.VenvDirLocator{}
		pipenvVenvDirLocator.LocateVenvDirCall.Returns.VenvDir = "some-venv-dir"

		// poetry
		poetryEntryResolver = &poetryfakes.EntryResolver{}
		poetryInstallProcess = &poetryfakes.InstallProcess{}
		poetryInstallProcess.ExecuteCall.Returns.String = "some-venv-dir"
		poetryPythonPathProcess = &poetryfakes.PythonPathLookupProcess{}
		poetryPythonPathProcess.ExecuteCall.Returns.String = "some-python-path"

		// uv
		uvRunner = &uvfakes.Runner{}

		buildParameters = build.CommonBuildParameters{
			SbomGenerator: sbomGenerator,
			Clock:         chronos.DefaultClock,
			Logger:        logger,
		}

		packagerParameters := map[string]pythonpackagers.PackagerParameters{
			conda.CondaEnvPlanEntry: conda.CondaBuildParameters{
				Runner: runner,
			},
			pipinstall.Manager: pipinstall.PipBuildParameters{
				InstallProcess:      pipInstallProcess,
				SitePackagesProcess: pipSitePackagesProcess,
			},
			pipenvinstall.Manager: pipenvinstall.PipEnvBuildParameters{
				InstallProcess: pipenvInstallProcess,
				SiteProcess:    pipenvSitePackagesProcess,
				VenvDirLocator: pipenvVenvDirLocator,
			},
			poetryinstall.PoetryVenv: poetryinstall.PoetryEnvBuildParameters{
				EntryResolver:           poetryEntryResolver,
				InstallProcess:          poetryInstallProcess,
				PythonPathLookupProcess: poetryPythonPathProcess,
			},
			uvinstall.UvEnvPlanEntry: uvinstall.UvBuildParameters{
				Runner: uvRunner,
			},
		}

		buildFunc = pythonpackagers.Build(logger, buildParameters, packagerParameters)

		buildContext = packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "Some Buildpack",
				Version:     "some-version",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
			},
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			// Plan: shall be filled within each test
			Platform: packit.Platform{Path: "some-platform-path"},
			Layers:   packit.Layers{Path: layersDir},
			Stack:    "some-stack",
		}

		plans = []packit.BuildpackPlan{
			packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: conda.CondaEnvPlanEntry,
					},
					{
						Name: pipinstall.Manager,
					},
					{
						Name: pipenvinstall.Manager,
					},
					{
						Name: poetryinstall.PoetryVenv,
					},
					{
						Name: uvinstall.UvEnvPlanEntry,
					},
				},
			},
			packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: conda.CondaEnvPlanEntry,
					},
					{
						Name: pipinstall.Manager,
					},
					{
						Name: pipenvinstall.Manager,
					},
				},
			},
			packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: conda.CondaEnvPlanEntry,
					},
					{
						Name: pipinstall.Manager,
					},
				},
			},
			packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: conda.CondaEnvPlanEntry,
					},
				},
			},
			packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: uvinstall.UvEnvPlanEntry,
					},
				},
			},
		}
	})

	it("runs the build process and returns expected layers", func() {
		for _, plan := range plans {
			buildContext.Plan = plan
			result, err := buildFunc(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(len(plan.Entries)))
		}
	})

	it("fails if packager parameters is missing", func() {
		packagerParameters := map[string]pythonpackagers.PackagerParameters{}

		buildFunc = pythonpackagers.Build(logger, buildParameters, packagerParameters)

		for _, plan := range plans {
			buildContext.Plan = plan
			_, err := buildFunc(buildContext)
			Expect(err).To(HaveOccurred())
		}
	})

	it("handles python-packagers opt-in", func() {
		for _, plan := range plans {
			buildContext.Plan = plan
			buildContext.Plan.Entries = append(buildContext.Plan.Entries,
				packit.BuildpackPlanEntry{
					Name: pythonpackagers.PackageManagersRunPlanEntry,
				},
			)
			result, err := buildFunc(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(len(plan.Entries)))
		}
	})
}
