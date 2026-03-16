// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	pythonpackagers "github.com/paketo-buildpacks/python-package-managers-run"
	pip "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pip"
	poetry "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/poetry"
	uv "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/uv"
)

func testPyProjectHandler(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "testdata")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Parse", func() {
		context("when backend is known", func() {
			it.Before(func() {
				content := []byte(`
				[build-system]
				requires = ["uv_build >= 0.9.28, <0.10.0"]
				build-backend = "uv_build"
				`)
				Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), content, 0755)).To(Succeed())
			})

			it("returns uv", func() {
				parser := pythonpackagers.NewPyProjectHandler()
				backend, err := parser.GetBuildBackend(workingDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(backend).To(Equal("uv_build"))

				installer, err := parser.GetInstaller(workingDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(installer).To(Equal("uv"))
			})

		})

		context("when the backend is empty", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte(""), 0755)).To(Succeed())
			})
			it.After(func() {

			})

			it("returns poetry", func() {
				parser := pythonpackagers.NewPyProjectHandler()
				backend, err := parser.GetBuildBackend(workingDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(backend).To(Equal(""))

				installer, err := parser.GetInstaller(workingDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(installer).To(Equal("poetry"))
			})
		})

		context("failure cases", func() {
			context("when the backend is unknown", func() {
				it.Before(func() {
					content := []byte(`
					[build-system]
					build-backend = "dummy"
					`)

					Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), content, 0755)).To(Succeed())
				})
				it.After(func() {

				})

				it("returns an error", func() {
					parser := pythonpackagers.NewPyProjectHandler()
					backend, err := parser.GetBuildBackend(workingDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(backend).To(Equal("dummy"))

					installer, err := parser.GetInstaller(workingDir)
					Expect(err).To(MatchError("unsupported backend: dummy"))
					Expect(installer).To(Equal(""))
				})
			})
		})
	})

	context("Create plan", func() {
		context("when the installer is known", func() {
			context("pip", func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(workingDir, "requirements.txt"), []byte(""), 0755)).To(Succeed())
				})
				it("creates a plan", func() {
					parser := pythonpackagers.NewPyProjectHandler()
					result, err := parser.Detect("pip", packit.DetectContext{
						WorkingDir: workingDir,
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(result.Plan.Provides[0].Name).To(Equal(pip.SitePackages))
				})
			})

			context("poetry", func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte(""), 0755)).To(Succeed())
				})
				it("creates a poetry plan", func() {
					parser := pythonpackagers.NewPyProjectHandler()
					result, err := parser.Detect("poetry", packit.DetectContext{
						WorkingDir: workingDir,
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(result.Plan.Provides[0].Name).To(Equal(poetry.PoetryVenv))
				})
			})

			context("uv", func() {
				context("with build-system entry and lock file", func() {
					it.Before(func() {
						content := []byte(`
					[build-system]
					requires = ["uv_build >= 0.9.28, <0.10.0"]
					build-backend = "uv_build"
					`)
						Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), content, 0755)).To(Succeed())
						Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), []byte(""), 0755)).To(Succeed())
					})
					it("creates a uv plan", func() {
						parser := pythonpackagers.NewPyProjectHandler()
						result, err := parser.Detect("uv", packit.DetectContext{
							WorkingDir: workingDir,
						})
						Expect(err).NotTo(HaveOccurred())
						Expect(result.Plan.Provides[0].Name).To(Equal(uv.UvEnvPlanEntry))
					})
				})
				context("without build-system entry but with a lock file", func() {
					it.Before(func() {
						Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte(""), 0755)).To(Succeed())
						Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), []byte(""), 0755)).To(Succeed())
					})
					it("creates a uv plan", func() {
						parser := pythonpackagers.NewPyProjectHandler()
						result, err := parser.Detect("uv", packit.DetectContext{
							WorkingDir: workingDir,
						})
						Expect(err).NotTo(HaveOccurred())
						Expect(result.Plan.Provides[0].Name).To(Equal(uv.UvEnvPlanEntry))
					})
				})
			})

		})
		context("failure cases", func() {
			it("fails when the installer is unknown", func() {
				parser := pythonpackagers.NewPyProjectHandler()
				_, err := parser.Detect("dummy", packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("unsupported installer: dummy"))
			})
		})
	})
}
