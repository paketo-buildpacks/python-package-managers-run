// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipenvinstall_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPipenvInstall(t *testing.T) {
	suite := spec.New("pipenvinstall", spec.Report(report.Terminal{}))
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite("InstallProcess", testInstallProcess)
	suite("LockParser", testLockParser)
	suite("PipfileParser", testPipfileParser)
	suite("SitePackagesProcess", testSiteProcess)
	suite("VenvLocator", testVenvLocator)
	suite.Run(t)
}
