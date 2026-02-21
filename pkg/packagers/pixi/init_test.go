// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixiinstall_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPixiInstall(t *testing.T) {
	suite := spec.New("pixiinstall", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("PixiRunner", testPixiRunner, spec.Sequential())
	suite("Detect", testDetect)
	suite.Run(t)
}
