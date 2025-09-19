// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetryinstall_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPoetryInstall(t *testing.T) {
	suite := spec.New("poetryinstall", spec.Report(report.Terminal{}))
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite("InstallProcess", testInstallProcess)
	suite("PythonPathProcess", testPythonPathProcess)
	suite.Run(t)
}
