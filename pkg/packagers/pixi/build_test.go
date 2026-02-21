// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixiinstall_test

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

	"github.com/paketo-buildpacks/python-packagers/pkg/build"
	pixiinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/pixi"
	"github.com/paketo-buildpacks/python-packagers/pkg/packagers/pixi/fakes"
	sbomfakes "github.com/paketo-buildpacks/python-packagers/pkg/sbom/fakes"

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
		sbomGenerator *sbomfakes.SBOMGenerator

		buildFunc    packit.BuildFunc
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
		sbomGenerator = &sbomfakes.SBOMGenerator{}

		runner.ShouldRunCall.Returns.Bool = true
		runner.ShouldRunCall.Returns.String = "some-sha"

		sbomGenerator.GenerateCall.Returns.SBOM = sbom.SBOM{}

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		buildFunc = pixiinstall.Build(
			pixiinstall.PixiBuildParameters{
				runner,
			},
			build.CommonBuildParameters{
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
						Name: pixiinstall.PixiEnvPlanEntry,
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
		result, err := buildFunc(buildContext)
		Expect(err).NotTo(HaveOccurred())

		layers := result.Layers
		Expect(layers).To(HaveLen(1))

		pixiEnvLayer := layers[0]
		Expect(pixiEnvLayer.Name).To(Equal("pixi-env"))
		Expect(pixiEnvLayer.Path).To(Equal(filepath.Join(layersDir, "pixi-env")))

		Expect(pixiEnvLayer.Build).To(BeFalse())
		Expect(pixiEnvLayer.Launch).To(BeFalse())
		Expect(pixiEnvLayer.Cache).To(BeFalse())

		Expect(pixiEnvLayer.BuildEnv).To(BeEmpty())
		Expect(pixiEnvLayer.LaunchEnv).To(BeEmpty())
		Expect(pixiEnvLayer.ProcessLaunchEnv).To(BeEmpty())

		Expect(pixiEnvLayer.SharedEnv).To(HaveLen(2))
		Expect(pixiEnvLayer.SharedEnv["PATH.prepend"]).To(Equal(filepath.Join(layersDir, "pixi-env", "default", "bin")))
		Expect(pixiEnvLayer.SharedEnv["PATH.delim"]).To(Equal(":"))

		Expect(pixiEnvLayer.Metadata).To(HaveLen(1))
		Expect(pixiEnvLayer.Metadata["lockfile-sha"]).To(Equal("some-sha"))

		Expect(pixiEnvLayer.SBOM.Formats()).To(HaveLen(2))
		var actualExtensions []string
		for _, format := range pixiEnvLayer.SBOM.Formats() {
			actualExtensions = append(actualExtensions, format.Extension)
		}
		Expect(actualExtensions).To(ConsistOf("cdx.json", "spdx.json"))

		Expect(runner.ExecuteCall.Receives.PixiEnvPath).To(Equal(filepath.Join(layersDir, "pixi-env")))
		Expect(runner.ExecuteCall.Receives.PixiCachePath).To(Equal(filepath.Join(layersDir, "pixi-env-cache")))
		Expect(runner.ExecuteCall.Receives.WorkingDir).To(Equal(workingDir))

		Expect(sbomGenerator.GenerateCall.Receives.Dir).To(Equal(filepath.Join(layersDir, "pixi-env")))
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
			Expect(os.RemoveAll(filepath.Join(layersDir, "pixi-env-cache"))).To(Succeed())
		})

		it("cache layer is exported", func() {
			result, err := buildFunc(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(2))

			pixiEnvLayer := layers[0]
			Expect(pixiEnvLayer.Name).To(Equal("pixi-env"))

			cacheLayer := layers[1]
			Expect(cacheLayer.Name).To(Equal("pixi-env-cache"))
			Expect(cacheLayer.Path).To(Equal(filepath.Join(layersDir, "pixi-env-cache")))

			Expect(cacheLayer.Build).To(BeFalse())
			Expect(cacheLayer.Launch).To(BeFalse())
			Expect(cacheLayer.Cache).To(BeTrue())
		})
	})

	context("when a build plan entry requires pixi-environment at launch", func() {
		it.Before(func() {
			buildContext.Plan.Entries[0].Metadata = map[string]interface{}{
				"launch": true,
			}
		})

		it("assigns the flag to the pixi env layer", func() {
			result, err := buildFunc(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(1))

			pixiEnvLayer := layers[0]
			Expect(pixiEnvLayer.Name).To(Equal("pixi-env"))

			Expect(pixiEnvLayer.Build).To(BeFalse())
			Expect(pixiEnvLayer.Launch).To(BeTrue())
			Expect(pixiEnvLayer.Cache).To(BeFalse())
		})
	})

	context("when a build plan entry requires pixi-environment at build", func() {
		it.Before(func() {
			buildContext.Plan.Entries[0].Metadata = map[string]interface{}{
				"build": true,
			}
		})

		it("assigns build and cache to the pixi env layer", func() {
			result, err := buildFunc(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(1))

			pixiEnvLayer := layers[0]
			Expect(pixiEnvLayer.Name).To(Equal("pixi-env"))

			Expect(pixiEnvLayer.Build).To(BeTrue())
			Expect(pixiEnvLayer.Launch).To(BeFalse())
			Expect(pixiEnvLayer.Cache).To(BeTrue())
		})
	})

	context("cached packages should be reused", func() {
		it.Before(func() {
			runner.ShouldRunCall.Returns.Bool = false
			runner.ShouldRunCall.Returns.String = "cached-sha"
		})

		it("reuses cached pixi env layer instead of running build process", func() {
			result, err := buildFunc(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(1))

			pixiEnvLayer := layers[0]
			Expect(pixiEnvLayer.Name).To(Equal("pixi-env"))

			Expect(runner.ExecuteCall.CallCount).To(BeZero())
		})
	})

	context("failure cases", func() {
		context("pixi layer cannot be fetched", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(layersDir, "pixi-env.toml"), nil, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := buildFunc(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("pixi cache layer cannot be fetched", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(layersDir, "pixi-env-cache.toml"), nil, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := buildFunc(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("runner ShouldRun fails", func() {
			it.Before(func() {
				runner.ShouldRunCall.Returns.Error = errors.New("some-shouldrun-error")
			})

			it("returns an error", func() {
				_, err := buildFunc(buildContext)
				Expect(err).To(MatchError("some-shouldrun-error"))
			})
		})

		context("layer cannot be reset", func() {
			it.Before(func() {
				runner.ShouldRunCall.Returns.Bool = true
				Expect(os.Chmod(layersDir, 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := buildFunc(buildContext)
				Expect(err).To(MatchError(ContainSubstring("error could not create directory")))
			})
		})

		context("install process fails to execute", func() {
			it.Before(func() {
				runner.ShouldRunCall.Returns.Bool = true
				runner.ExecuteCall.Returns.Error = errors.New("some execution error")
			})

			it("returns an error", func() {
				_, err := buildFunc(buildContext)
				Expect(err).To(MatchError(ContainSubstring("some execution error")))
			})
		})

		context("when generating the SBOM returns an error", func() {
			it.Before(func() {
				buildContext.BuildpackInfo.SBOMFormats = []string{"random-format"}
			})

			it("returns an error", func() {
				_, err := buildFunc(buildContext)
				Expect(err).To(MatchError(`unsupported SBOM format: 'random-format'`))
			})
		})

		context("when formatting the SBOM returns an error", func() {
			it.Before(func() {
				sbomGenerator.GenerateCall.Returns.Error = errors.New("failed to generate SBOM")
			})

			it("returns an error", func() {
				_, err := buildFunc(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to generate SBOM")))
			})
		})
	})
}
