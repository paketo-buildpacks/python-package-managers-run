// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPythonPackagers(t *testing.T) {
	suite := spec.New("python-packagers", spec.Report(report.Terminal{}), spec.Sequential())
	suite("PyProjectHandler", testPyProjectHandler)
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite.Run(t)
}
