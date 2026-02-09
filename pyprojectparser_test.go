// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	pythonpackagers "github.com/paketo-buildpacks/python-packagers"
)

func testPyProjectParser(t *testing.T, context spec.G, it spec.S) {
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
				parser := pythonpackagers.NewPyProjectParser()
				pyproject := filepath.Join(workingDir, "pyproject.toml")
				backend, err := parser.GetBuildBackend(pyproject)
				Expect(err).NotTo(HaveOccurred())
				Expect(backend).To(Equal("uv_build"))

				installer, err := parser.GetInstaller(pyproject)
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
				parser := pythonpackagers.NewPyProjectParser()
				pyproject := filepath.Join(workingDir, "pyproject.toml")
				backend, err := parser.GetBuildBackend(pyproject)
				Expect(err).NotTo(HaveOccurred())
				Expect(backend).To(Equal(""))

				installer, err := parser.GetInstaller(pyproject)
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
					parser := pythonpackagers.NewPyProjectParser()
					pyproject := filepath.Join(workingDir, "pyproject.toml")
					backend, err := parser.GetBuildBackend(pyproject)
					Expect(err).NotTo(HaveOccurred())
					Expect(backend).To(Equal("dummy"))

					installer, err := parser.GetInstaller(pyproject)
					Expect(err).To(MatchError("unsupported backend: dummy"))
					Expect(installer).To(Equal(""))
				})
			})
		})
	})

	context("Create plan", func() {
		context("when the installer is known", func() {
			it("creates a valid plan", func() {
				parser := pythonpackagers.NewPyProjectParser()
				plan, err := parser.CreatePlan("pip")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Provides[0].Name).To(Equal("site-packages"))
			})
		})
		context("failure cases", func() {
			it("fails when the installer is unknown", func() {
				parser := pythonpackagers.NewPyProjectParser()
				_, err := parser.CreatePlan("dummy")
				Expect(err).To(MatchError("unsupported installer: dummy"))
			})
		})
	})
}
