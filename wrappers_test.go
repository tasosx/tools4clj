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
	"os/exec"
	"runtime"
	"testing"
)

func TestRlwrapArgs(t *testing.T) {
	expected := []string{
		"rlwrap",
		"-r",
		"-q",
		`"`,
		"-b",
		`(){}[],^%#@";:'`,
	}
	args := rlwrapArgs()

	if len(args) != len(expected) {
		t.Errorf("rebelArgs failed, args expected %v, got %v", len(expected), len(args))
		t.FailNow()
	}

	for i, v := range args {
		if v != expected[i] {
			t.Errorf("rebelArgs failed, arg expected %v, got %v", expected[i], v)
		}
	}
}

func TestRlwrapPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("on windows platform, skip 'rlwrap path' test")
	}

	path := rlwrapPath()

	expected, err := exec.LookPath("rlwrap")
	if err != nil {
		t.Errorf("rlwrap path failed, with error %v", err)
	}
	if path != expected {
		t.Errorf("rlwrap failed, arg expected %v, got %v", expected, path)
	}
}
