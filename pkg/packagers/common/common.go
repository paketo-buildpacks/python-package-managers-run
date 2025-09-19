// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers

import (
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type SBOMGenerator interface {
	Generate(dir string) (sbom.SBOM, error)
}

type Generator struct{}

func (f Generator) Generate(dir string) (sbom.SBOM, error) {
	return sbom.Generate(dir)
}

// CommonBuildParameters are the parameters shared
// by all packager build function implementation
type CommonBuildParameters struct {
	SbomGenerator SBOMGenerator
	Clock         chronos.Clock
	Logger        scribe.Emitter
}
