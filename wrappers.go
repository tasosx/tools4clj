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

import "os/exec"

const (
	rebelSdepsArg = `{:deps {com.bhauman/rebel-readline {:mvn/version "0.1.4"}}}`
	rebelMainArg  = "rebel-readline.main"
)

func rlwrapArgs() []string {
	return []string{
		"rlwrap",
		"-r",
		"-q",
		`"`,
		"-b",
		`(){}[],^%#@";:'`,
	}
}

func rlwrapPath() string {
	path, err := exec.LookPath("rlwrap")
	if err != nil {
		return ""
	}
	return path
}
