// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythonpackagers

import (
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	conda "github.com/paketo-buildpacks/python-packagers/pkg/packagers/conda"
	pipinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/pip"
	pipenvinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/pipenv"
	poetryinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/poetry"
	uvinstall "github.com/paketo-buildpacks/python-packagers/pkg/packagers/uv"

	pythonpackagers "github.com/paketo-buildpacks/python-packagers/pkg/packagers/common"
)

// filtered returns the slice passed in parameter with the needle removed
func filtered(haystack []packit.BuildpackPlanEntry, needle string) []packit.BuildpackPlanEntry {
	output := []packit.BuildpackPlanEntry{}

	for _, entry := range haystack {
		if entry.Name != needle {
			output = append(output, entry)
		}
	}

	return output
}

type PackagerParameters interface {
}

func Build(
	logger scribe.Emitter,
	commonBuildParameters pythonpackagers.CommonBuildParameters,
	buildParameters map[string]PackagerParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		planEntries := filtered(context.Plan.Entries, pipinstall.SitePackages)
		layers := []packit.Layer{}

		for _, entry := range planEntries {
			logger.Title("Handling %s", entry.Name)

			switch entry.Name {
			case pipinstall.Manager:
				if parameters, ok := buildParameters[pipinstall.Manager]; ok {
					pipResult, err := pipinstall.Build(
						parameters.(pipinstall.PipBuildParameters),
						commonBuildParameters,
					)(context)

					if err != nil {
						return packit.BuildResult{}, err
					}

					layers = append(layers, pipResult.Layers...)
				} else {
					return packit.BuildResult{}, packit.Fail.WithMessage("missing plan for: %s", entry.Name)
				}

			case pipenvinstall.Manager:
				if parameters, ok := buildParameters[pipenvinstall.Manager]; ok {
					pipEnvResult, err := pipenvinstall.Build(
						parameters.(pipenvinstall.PipEnvBuildParameters),
						commonBuildParameters,
					)(context)

					if err != nil {
						return packit.BuildResult{}, err
					}

					layers = append(layers, pipEnvResult.Layers...)
				} else {
					return packit.BuildResult{}, packit.Fail.WithMessage("missing plan for: %s", entry.Name)
				}
			case conda.CondaEnvPlanEntry:
				if parameters, ok := buildParameters[conda.CondaEnvPlanEntry]; ok {
					condaResult, err := conda.Build(
						parameters.(conda.CondaBuildParameters),
						commonBuildParameters,
					)(context)

					if err != nil {
						return packit.BuildResult{}, err
					}

					layers = append(layers, condaResult.Layers...)
				} else {
					return packit.BuildResult{}, packit.Fail.WithMessage("missing plan for: %s", entry.Name)
				}
			case poetryinstall.PoetryVenv:
				if parameters, ok := buildParameters[poetryinstall.PoetryVenv]; ok {
					poetryResult, err := poetryinstall.Build(
						parameters.(poetryinstall.PoetryEnvBuildParameters),
						commonBuildParameters,
					)(context)

					if err != nil {
						return packit.BuildResult{}, err
					}

					layers = append(layers, poetryResult.Layers...)
				} else {
					return packit.BuildResult{}, packit.Fail.WithMessage("missing plan for: %s", entry.Name)
				}
			case uvinstall.UvEnvPlanEntry:
				if parameters, ok := buildParameters[uvinstall.UvEnvPlanEntry]; ok {
					uvResult, err := uvinstall.Build(
						parameters.(uvinstall.UvBuildParameters),
						commonBuildParameters,
					)(context)

					if err != nil {
						return packit.BuildResult{}, err
					}

					layers = append(layers, uvResult.Layers...)
				} else {
					return packit.BuildResult{}, packit.Fail.WithMessage("missing plan for: %s", entry.Name)
				}
			default:
				return packit.BuildResult{}, packit.Fail.WithMessage("unknown plan: %s", entry.Name)
			}
		}

		if len(layers) == 0 {
			return packit.BuildResult{}, packit.Fail.WithMessage("empty plan should not happen")
		}

		return packit.BuildResult{
			Layers: layers,
		}, nil
	}
}
