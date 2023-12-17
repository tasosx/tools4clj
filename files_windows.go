//go:build windows

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
	"path"
)

func isReadOnlyDir(pathname string) bool {
	_, err := os.Stat(pathname)
	if os.IsNotExist(err) {
		return true
	}

	filepath := path.Join(pathname, "tmp.tools4clj.isRO")

	// try to create a tmp file
	out, err := os.Create(filepath)
	if err != nil {
		// on a permission error its a read-only dir
		return os.IsPermission(err)
	} else {
		defer os.Remove(filepath)
		defer out.Close()
	}
	return false
}
