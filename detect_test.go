// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	pythonpackagers "github.com/paketo-buildpacks/python-package-managers-run"
	"github.com/paketo-buildpacks/python-package-managers-run/pkg/build"
	conda "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/conda"
	pip "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pip"
	pipenv "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pipenv"
	poetry "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/poetry"
	uv "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/uv"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		buffer     *bytes.Buffer

		detect packit.DetectFunc
	)

	it.Before(func() {
		workingDir = t.TempDir()

		Expect(os.WriteFile(filepath.Join(workingDir, "x.py"), []byte{}, os.ModePerm)).To(Succeed())

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		detect = pythonpackagers.Detect(logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("detection phase", func() {
		context("When only an environment.yml file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "environment.yml"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: conda.CondaEnvPlanEntry,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: conda.CondaPlanEntry,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))

			})
		})

		context("When only a package-list.txt file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "package-list.txt"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: conda.CondaEnvPlanEntry,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: conda.CondaPlanEntry,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))
			})
		})

		context("When only a requirements.txt file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "requirements.txt"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
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
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pip.Pip,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pip.Manager,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))
			})
		})

		context("When only a Pipfile file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "Pipfile"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: pipenv.SitePackages,
						},
						{
							Name: pipenv.Manager,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: pipenv.CPython,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pipenv.Pipenv,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pipenv.Manager,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))
			})
		})

		context("When only a pyproject.toml file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: poetry.PoetryVenv,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: poetry.CPython,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: poetry.Poetry,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))
			})
		})

		context("When a uv.lock file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "uv.lock"), []byte{}, os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: uv.UvEnvPlanEntry,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: uv.UvPlanEntry,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))
			})
		})

		context("When a uv.lock and pyproject.toml file is present", func() {
			it.Before(func() {
				content := []byte(`
					[build-system]
					requires = ["uv_build>=0.10.0,<0.11.0"]
					build-backend = "uv_build"
				`)
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), content, os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "uv.lock"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: uv.UvEnvPlanEntry,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: uv.UvPlanEntry,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))
			})
		})

		context("When no python related files are present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("No python packager manager related files found")))
			})
		})
	})
}
