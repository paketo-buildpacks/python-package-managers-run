// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go

// Executable defines the interface for invoking an executable.
type Executable interface {
	Execute(pexec.Execution) error
}

// Summer defines the interface for computing a SHA256 for a set of files
// and/or directories.
//
//go:generate faux --interface Summer --output fakes/summer.go
type Summer interface {
	Sum(arg ...string) (string, error)
}

// UvRunner implements the Runner interface.
type UvRunner struct {
	executable Executable
	summer     Summer
	logger     scribe.Emitter
}

// NewUvRunner creates an instance of UvRunner given an Executable and a Logger.
func NewUvRunner(executable Executable, summer Summer, logger scribe.Emitter) UvRunner {
	return UvRunner{
		executable: executable,
		summer:     summer,
		logger:     logger,
	}
}

// ShouldRun determines whether the uv environment setup command needs to be
// run, given the path to the app directory and the metadata from the
// preexisting uv-env layer. It returns true if the uv environment setup
// command must be run during this build, the SHA256 of the uv.lock in
// the app directory, and an error. If there is no uv.lock, the sha
// returned is an empty string.
func (c UvRunner) ShouldRun(workingDir string, metadata map[string]interface{}) (run bool, sha string, err error) {
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

// Execute runs the uv environment setup command and cleans up unnecessary
// artifacts. If a vendor directory is present, it uses vendored packages and
// installs them in offline mode. In this case it will use the cpython layer
// to create the virtual environment for the installation
func (c UvRunner) Execute(uvLayerPath string, uvCachePath string, workingDir string) error {
	lockfileExists, err := fs.Exists(filepath.Join(workingDir, LockfileName))
	if err != nil {
		return err
	}

	if !lockfileExists {
		return errors.New("missing lock file")
	}

	vendorDir := filepath.Join(workingDir, "vendor")

	exists, err := fs.Exists(vendorDir)
	if err != nil {
		return err
	}

	venvPath := filepath.Join(uvLayerPath, "venv")
	args := []string{
		"venv",
		venvPath,
	}

	env := append(os.Environ(), fmt.Sprintf("HOME=%s", uvLayerPath))

	if exists {
		args = append(args, "--offline", "--python", "/layers/paketo-buildpacks_cpython/cpython/bin/python")
		env = append(env, "LD_LIBRARY_PATH=/layers/paketo-buildpacks_cpython/cpython/lib")
	}

	c.logger.Subprocess("Running 'uv %s'", strings.Join(args, " "))

	err = c.executable.Execute(pexec.Execution{
		Args:   args,
		Env:    env,
		Stdout: c.logger.ActionWriter,
		Stderr: c.logger.ActionWriter,
	})

	if err != nil {
		return fmt.Errorf("failed to run uv command: %w", err)
	}

	userFindLinks, _ := os.LookupEnv("BP_UV_FIND_LINKS")
	findLinks, _ := os.LookupEnv("UV_FIND_LINKS")

	combinedFindLinks := []string{userFindLinks, findLinks}

	if exists {
		combinedFindLinks = append(combinedFindLinks, vendorDir)
		args = offlineArgs(venvPath, workingDir)
	} else {
		args = onlineArgs(venvPath, uvCachePath, workingDir)
	}

	c.logger.Subprocess("Running 'uv %s'", strings.Join(args, " "))

	err = c.executable.Execute(pexec.Execution{
		Args: args,
		Env: append(env,
			fmt.Sprintf("UV_FIND_LINKS=%s", strings.TrimLeft(strings.Join(combinedFindLinks, " "), " ")),
		),

		Stdout: c.logger.ActionWriter,
		Stderr: c.logger.ActionWriter,
	})

	if err != nil {
		return fmt.Errorf("failed to run uv command: %w", err)
	}

	return nil
}

func onlineArgs(venvPath string, cachePath string, workingDir string) []string {
	return []string{
		"pip",
		"install",
		"--python",
		filepath.Join(venvPath, "bin", "python"),
		"--cache-dir",
		cachePath,
		workingDir,
	}
}

func offlineArgs(venvPath string, workingDir string) []string {
	return []string{
		"pip",
		"install",
		"--no-index",
		"--python",
		filepath.Join(venvPath, "bin", "python"),
		workingDir,
		"--offline",
	}
}
