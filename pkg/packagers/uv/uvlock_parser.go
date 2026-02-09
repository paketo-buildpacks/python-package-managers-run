// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall

import (
	"strings"

	"github.com/BurntSushi/toml"
)

type Lockfile struct {
	RequiresPython string `toml:"requires-python"`
}

type LockfileParser struct {
}

func NewLockfileParser() LockfileParser {
	return LockfileParser{}
}

func (p LockfileParser) ParsePythonVersion(lockfilePath string) (string, error) {
	var lockfile Lockfile

	_, err := toml.DecodeFile(lockfilePath, &lockfile)
	if err != nil {
		return "", err
	}

	if lockfile.RequiresPython != "" {
		return strings.Trim(lockfile.RequiresPython, "="), nil
	}
	return lockfile.RequiresPython, nil
}
