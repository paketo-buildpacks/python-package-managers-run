// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixiinstall_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	common "github.com/paketo-buildpacks/python-packagers/pkg/packagers/common"
	pixiinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/pixi"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		detect     packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		detect = pixiinstall.Detect()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there is an pixi.lock in the working dir", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, pixiinstall.LockfileName), nil, 0644)).To(Succeed())
		})

		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{
						Name: "pixi-environment",
					},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "pixi",
						Metadata: common.BuildPlanMetadata{
							Build: true,
						},
					},
				},
			}))
		})
	})

	context("when there is an pixi.toml in the working dir", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, pixiinstall.ProjectFilename), nil, 0644)).To(Succeed())
		})

		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{
						Name: "pixi-environment",
					},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "pixi",
						Metadata: common.BuildPlanMetadata{
							Build: true,
						},
					},
				},
			}))
		})
	})

	context("when no pixi.toml or pixi.lock is present in the working dir", func() {
		it("fails to detect", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail.WithMessage("no 'pixi.toml' and 'pixi.lock' found")))
		})
	})

	context("failure cases", func() {
		context("when the file cannot be stat'd", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("failed trying to stat pixi.toml:")))
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
