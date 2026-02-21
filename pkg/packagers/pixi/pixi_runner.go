// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixiinstall

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	"github.com/paketo-buildpacks/python-packagers/pkg/executable"
)

// Summer defines the interface for computing a SHA256 for a set of files
// and/or directories.
//
//go:generate faux --interface Summer --output fakes/summer.go
type Summer interface {
	Sum(arg ...string) (string, error)
}

// PixiRunner implements the Runner interface.
type PixiRunner struct {
	logger     scribe.Emitter
	executable executable.Executable
	summer     Summer
}

type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Build   string `json:"build"`
}

// NewPixiRunner creates an instance of PixiRunner given an Executable, a Summer, and a Logger.
func NewPixiRunner(executable executable.Executable, summer Summer, logger scribe.Emitter) PixiRunner {
	return PixiRunner{
		executable: executable,
		summer:     summer,
		logger:     logger,
	}
}

// ShouldRun determines whether the pixi environment setup command needs to be
// run, given the path to the app directory and the metadata from the
// preexisting pixi-env layer. It returns true if the pixi environment setup
// command must be run during this build, the SHA256 of the package-list.txt in
// the app directory, and an error. If there is no package-list.txt, the sha
// returned is an empty string.
func (c PixiRunner) ShouldRun(workingDir string, metadata map[string]interface{}) (run bool, sha string, err error) {
	lockfilePath := filepath.Join(workingDir, LockfileName)
	_, err = os.Stat(lockfilePath)

	if errors.Is(err, os.ErrNotExist) {
		return true, "", nil
	}

	if err != nil {
		return false, "", err
	}

	updatedLockfileSha, err := c.summer.Sum(lockfilePath)
	if err != nil {
		return false, "", err
	}

	if updatedLockfileSha == metadata[LockfileShaName] {
		return false, updatedLockfileSha, nil
	}

	return true, updatedLockfileSha, nil
}

// Do the actual command execution
func (c PixiRunner) run(args []string) error {
	c.logger.Subprocess("Running 'pixi %s'", strings.Join(args, " "))

	err := c.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: c.logger.ActionWriter,
		Stderr: c.logger.ActionWriter,
	})

	if err != nil {
		return fmt.Errorf("failed to run pixi command: %w", err)
	}

	return err
}

// Execute runs the pixi pack and unpack command to create a usable
// environment.
//
// For more information about the commands used, see:
// https://pixi.prefix.dev/latest/deployment/pixi_pack/
func (c PixiRunner) Execute(pixiLayerPath string, pixiCachePath string, workingDir string) error {
	lockfileExists, err := fs.Exists(filepath.Join(workingDir, LockfileName))
	if err != nil {
		return err
	}
	projectFileExists, err := fs.Exists(filepath.Join(workingDir, ProjectFilename))
	if err != nil {
		return err
	}

	if !lockfileExists && !projectFileExists {
		return fmt.Errorf("missing both %s and %s", LockfileName, ProjectFilename)
	}

	args := []string{
		"exec",
		"pixi-pack",
		"--use-cache", pixiCachePath,
		"--output-file", "/tmp/project.tar.gz",
		workingDir,
	}

	err = c.run(args)

	if err != nil {
		return err
	}

	args = []string{
		"exec",
		"pixi-unpack",
		"--output-directory", pixiLayerPath,
		"--env-name", PixiEnvironmentName,
		"/tmp/project.tar.gz",
	}

	return c.run(args)
}
