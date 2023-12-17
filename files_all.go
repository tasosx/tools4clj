//go:build !windows

/*************************************************************************
 * Copyright (c) 2019 Tasos Mamaloukos.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution.
 *
 * The Eclipse Public License is available at
 *     https://www.eclipse.org/org/documents/epl-v10.html
 *
 *************************************************************************/

package tools4clj

import (
	"os"
)

func isReadOnlyDir(pathname string) bool {
	info, err := os.Stat(pathname)
	if os.IsNotExist(err) {
		return true
	}
	return (info.Mode() & os.ModePerm) != os.ModePerm
}
