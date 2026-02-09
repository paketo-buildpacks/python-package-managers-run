// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	pythonpackagers "github.com/paketo-buildpacks/python-packagers/pkg/packagers/common"
	uvinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/uv"
	"github.com/paketo-buildpacks/python-packagers/pkg/packagers/uv/fakes"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string

		buffer *bytes.Buffer

		runner        *fakes.Runner
		sbomGenerator *fakes.SBOMGenerator

		build        packit.BuildFunc
		buildContext packit.BuildContext
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		runner = &fakes.Runner{}
		sbomGenerator = &fakes.SBOMGenerator{}

		runner.ShouldRunCall.Returns.Bool = true
		runner.ShouldRunCall.Returns.String = "some-sha"

		sbomGenerator.GenerateCall.Returns.SBOM = sbom.SBOM{}

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		build = uvinstall.Build(
			uvinstall.UvBuildParameters{
				runner,
			},
			pythonpackagers.CommonBuildParameters{
				SbomGenerator: sbomGenerator,
				Clock:         chronos.DefaultClock,
				Logger:        logger,
			},
		)
		buildContext = packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "Some Buildpack",
				Version:     "some-version",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
			},
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: uvinstall.UvEnvPlanEntry,
					},
				},
			},
			Platform: packit.Platform{Path: "some-platform-path"},
			Layers:   packit.Layers{Path: layersDir},
			Stack:    "some-stack",
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that builds correctly", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		layers := result.Layers
		Expect(layers).To(HaveLen(1))

		uvEnvLayer := layers[0]
		Expect(uvEnvLayer.Name).To(Equal("uv-env"))
		Expect(uvEnvLayer.Path).To(Equal(filepath.Join(layersDir, "uv-env")))

		Expect(uvEnvLayer.Build).To(BeFalse())
		Expect(uvEnvLayer.Launch).To(BeFalse())
		Expect(uvEnvLayer.Cache).To(BeFalse())

		Expect(uvEnvLayer.BuildEnv).To(BeEmpty())
		Expect(uvEnvLayer.LaunchEnv).To(BeEmpty())
		Expect(uvEnvLayer.ProcessLaunchEnv).To(BeEmpty())
		// Expect(uvEnvLayer.SharedEnv).ToNot(BeEmpty())
		Expect(uvEnvLayer.SharedEnv).To(HaveLen(2))
		Expect(uvEnvLayer.SharedEnv["PATH.prepend"]).To(Equal(filepath.Join(uvEnvLayer.Path, "venv", "bin")))
		Expect(uvEnvLayer.SharedEnv["PATH.delim"]).To(Equal(":"))

		Expect(uvEnvLayer.SBOM.Formats()).To(HaveLen(2))
		var actualExtensions []string
		for _, format := range uvEnvLayer.SBOM.Formats() {
			actualExtensions = append(actualExtensions, format.Extension)
		}
		Expect(actualExtensions).To(ConsistOf("cdx.json", "spdx.json"))

		Expect(runner.ExecuteCall.Receives.UvEnvPath).To(Equal(filepath.Join(layersDir, "uv-env")))
		Expect(runner.ExecuteCall.Receives.UvCachePath).To(Equal(filepath.Join(layersDir, "uv-env-cache")))
		Expect(runner.ExecuteCall.Receives.WorkingDir).To(Equal(workingDir))

		Expect(sbomGenerator.GenerateCall.Receives.Dir).To(Equal(workingDir))
	})

	context("when the runner executes outputting a non-empty cache dir", func() {
		it.Before(func() {
			runner.ExecuteCall.Stub = func(_, c, _ string) error {
				Expect(os.Mkdir(c, os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(c, "some-file"), []byte{}, os.ModePerm)).To(Succeed())
				return nil
			}
		})

		it.After(func() {
			Expect(os.RemoveAll(filepath.Join(layersDir, "uv-env-cache"))).To(Succeed())
		})

		it("cache layer is exported", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(2))

			uvEnvLayer := layers[0]
			Expect(uvEnvLayer.Name).To(Equal("uv-env"))

			cacheLayer := layers[1]
			Expect(cacheLayer.Name).To(Equal("uv-env-cache"))
			Expect(cacheLayer.Path).To(Equal(filepath.Join(layersDir, "uv-env-cache")))

			Expect(cacheLayer.Build).To(BeFalse())
			Expect(cacheLayer.Launch).To(BeFalse())
			Expect(cacheLayer.Cache).To(BeTrue())
		})
	})

	context("when a build plan entry requires uv-environment at launch", func() {
		it.Before(func() {
			buildContext.Plan.Entries[0].Metadata = map[string]interface{}{
				"launch": true,
			}
		})

		it("assigns the flag to the uv env layer", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(1))

			uvEnvLayer := layers[0]
			Expect(uvEnvLayer.Name).To(Equal("uv-env"))

			Expect(uvEnvLayer.Build).To(BeFalse())
			Expect(uvEnvLayer.Launch).To(BeTrue())
			Expect(uvEnvLayer.Cache).To(BeFalse())
		})
	})

	context("when a build plan entry requires uv-environment at build", func() {
		it.Before(func() {
			buildContext.Plan.Entries[0].Metadata = map[string]interface{}{
				"build": true,
			}
		})

		it("assigns build and cache to the uv env layer", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(1))

			uvEnvLayer := layers[0]
			Expect(uvEnvLayer.Name).To(Equal("uv-env"))

			Expect(uvEnvLayer.Build).To(BeTrue())
			Expect(uvEnvLayer.Launch).To(BeFalse())
			Expect(uvEnvLayer.Cache).To(BeTrue())
		})
	})

	context("cached packages should be reused", func() {
		it.Before(func() {
			runner.ShouldRunCall.Returns.Bool = false
			runner.ShouldRunCall.Returns.String = "cached-sha"
		})

		it("reuses cached uv env layer instead of running build process", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(1))

			uvEnvLayer := layers[0]
			Expect(uvEnvLayer.Name).To(Equal("uv-env"))

			Expect(runner.ExecuteCall.CallCount).To(BeZero())
		})
	})

	context("failure cases", func() {
		context("uv layer cannot be fetched", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(layersDir, "uv-env.toml"), nil, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("uv cache layer cannot be fetched", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(layersDir, "uv-env-cache.toml"), nil, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("runner ShouldRun fails", func() {
			it.Before(func() {
				runner.ShouldRunCall.Returns.Error = errors.New("some-shouldrun-error")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("some-shouldrun-error"))
			})
		})

		context("layer cannot be reset", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("error could not create directory")))
			})
		})

		context("install process fails to execute", func() {
			it.Before(func() {
				runner.ShouldRunCall.Returns.Bool = true
				runner.ExecuteCall.Returns.Error = errors.New("some execution error")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("some execution error")))
			})
		})

		context("when generating the SBOM returns an error", func() {
			it.Before(func() {
				buildContext.BuildpackInfo.SBOMFormats = []string{"random-format"}
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(`unsupported SBOM format: 'random-format'`))
			})
		})

		context("when formatting the SBOM returns an error", func() {
			it.Before(func() {
				sbomGenerator.GenerateCall.Returns.Error = errors.New("failed to generate SBOM")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to generate SBOM")))
			})
		})
	})
}
