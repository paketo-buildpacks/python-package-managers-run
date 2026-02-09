// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitUvInstall(t *testing.T) {
	suite := spec.New("uvinstall", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("UvRunner", testUvRunner)
	suite("UvLockParser", testUvLockParser)
	suite("Detect", testDetect)
	suite.Run(t)
}
