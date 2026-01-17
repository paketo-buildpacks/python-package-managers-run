// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
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
	suite("Detect", testDetect)
	suite.Run(t)
}
