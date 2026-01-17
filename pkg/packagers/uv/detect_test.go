// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	uvinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/uv"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
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

		detect = uvinstall.Detect()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there is an uv.lock in the working dir", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, uvinstall.LockfileName), nil, 0644)).To(Succeed())
		})

		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{
						Name: uvinstall.UvEnvPlanEntry,
					},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: uvinstall.UvPlanEntry,
						Metadata: map[string]interface{}{
							"build": true,
						},
					},
				},
			}))
		})
	})

	context("when no uv.lock is present in the working dir", func() {
		it("fails to detect", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail.WithMessage("no 'uv.lock' found")))
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
				Expect(err).To(MatchError(ContainSubstring("failed trying to stat uv.lock:")))
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
