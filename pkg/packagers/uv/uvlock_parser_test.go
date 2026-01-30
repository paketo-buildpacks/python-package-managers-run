// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/python-packagers/pkg/packagers/uv"
	"github.com/sclevine/spec"
)

func testUvLockParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		lockfile   string

		parser uvinstall.LockfileParser
	)

	const (
		version = `requires-python = "==1.2.3"`
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		lockfile = filepath.Join(workingDir, uvinstall.LockfileName)

		parser = uvinstall.NewLockfileParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Calling ParsePythonVersion", func() {
		it("parses version", func() {
			Expect(os.WriteFile(lockfile, []byte(version), 0644)).To(Succeed())

			version, err := parser.ParsePythonVersion(lockfile)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("1.2.3"))
		})

		it("returns empty string if file does not contain requires-python", func() {
			Expect(os.WriteFile(lockfile, []byte(""), 0644)).To(Succeed())

			version, err := parser.ParsePythonVersion(lockfile)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal(""))
		})

		context("error handling", func() {
			it("fails if file does not exist", func() {
				_, err := parser.ParsePythonVersion("not-a-valid-dir")
				Expect(err).To(HaveOccurred())
			})
		})
	})
}
