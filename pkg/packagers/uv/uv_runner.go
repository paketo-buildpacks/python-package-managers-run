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

	"github.com/paketo-buildpacks/python-package-managers-run/pkg/executable"
	"github.com/paketo-buildpacks/python-package-managers-run/pkg/summer"
)

// UvRunner implements the Runner interface.
type UvRunner struct {
	executable executable.Executable
	summer     summer.Summer
	logger     scribe.Emitter
}

// NewUvRunner creates an instance of UvRunner given an Executable and a Logger.
func NewUvRunner(executable executable.Executable, summer summer.Summer, logger scribe.Emitter) UvRunner {
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

	venvPath := filepath.Join(uvLayerPath, "venv")

	userFindLinks, _ := os.LookupEnv("BP_UV_FIND_LINKS")
	findLinks, _ := os.LookupEnv("UV_FIND_LINKS")

	combinedFindLinks := []string{userFindLinks, findLinks}

	env := append(os.Environ(), fmt.Sprintf("HOME=%s", uvLayerPath))
	env = append(env, fmt.Sprintf("VIRTUAL_ENV=%s", venvPath))
	env = append(env, fmt.Sprintf("UV_PROJECT_ENVIRONMENT=%s", venvPath))
	env = append(env, fmt.Sprintf("UV_WORKING_DIR=%s", workingDir))

	args := []string{
		"sync",
	}
	vendorDir := filepath.Join(workingDir, "vendor")

	vendorExists, err := fs.Exists(vendorDir)
	if err != nil {
		return err
	}
	if vendorExists {
		env = append(env, "LD_LIBRARY_PATH=/layers/paketo-buildpacks_cpython/cpython/lib")
		env = append(env, "UV_OFFLINE=1")
		env = append(env, "UV_PYTHON=/layers/paketo-buildpacks_cpython/cpython/bin/python")
		combinedFindLinks = append(combinedFindLinks, vendorDir)
		args = append(args, "--no-index")
	} else {
		env = append(env, fmt.Sprintf("UV_CACHE_DIR=%s", uvCachePath))
	}
	env = append(env, fmt.Sprintf("UV_FIND_LINKS=%s", strings.TrimLeft(strings.Join(combinedFindLinks, " "), " ")))

	installGroups, installGroupsPresent := os.LookupEnv("BP_UV_INSTALL_GROUPS")
	if installGroupsPresent {
		for _, group := range strings.Split(installGroups, ",") {
			args = append(args, fmt.Sprintf("--group=%s", group))
		}
	}

	c.logger.Subprocess("%s\nRunning 'uv %s'", strings.Join(env, "\n"), strings.Join(args, " "))

	err = c.executable.Execute(pexec.Execution{
		Args: args,
		Env:  env,

		Stdout: c.logger.ActionWriter,
		Stderr: c.logger.ActionWriter,
	})

	if err != nil {
		return fmt.Errorf("failed to run uv command: %w", err)
	}

	return nil
}
