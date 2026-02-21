// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package sbom

import (
	"github.com/paketo-buildpacks/packit/v2/sbom"
)

//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

type SBOMGenerator interface {
	Generate(dir string) (sbom.SBOM, error)
}

type Generator struct{}

func (f Generator) Generate(dir string) (sbom.SBOM, error) {
	return sbom.Generate(dir)
}
