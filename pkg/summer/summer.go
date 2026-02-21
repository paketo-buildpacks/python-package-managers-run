// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package summer

// Summer defines the interface for computing a SHA256 for a set of files
// and/or directories.
//
//go:generate faux --interface Summer --output fakes/summer.go
type Summer interface {
	Sum(arg ...string) (string, error)
}
