// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func uvTestReused(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			images = map[string]bool{}
			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			for id := range images {
				Expect(docker.Image.Remove.Execute(id)).To(Succeed())
			}
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("reuses layers", func() {
			var err error
			var logs1 fmt.Stringer
			var logs2 fmt.Stringer

			source, err = occam.Source(filepath.Join("testdata", "uv", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			image1, logs1, err := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.PythonInstallers.Online,
					settings.Buildpacks.PythonPackagers.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs1.String)
			images[image1.ID] = true

			image2, logs2, err := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.PythonInstallers.Online,
					settings.Buildpacks.PythonPackagers.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs2.String)
			images[image2.ID] = true

			Expect(logs2).To(ContainLines(
				fmt.Sprintf("  Reusing cached layer /layers/%s/uv-env", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
			))

			Expect(image2.Buildpacks[0].Layers["uv-env"].SHA).To(Equal(image1.Buildpacks[0].Layers["uv-env"].SHA))
		})
	})
}
