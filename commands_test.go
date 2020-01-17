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

func TestResolveTagsCmd(t *testing.T) {
	toolsCpDir := "test-cpTools-dir"
	cmd := resolveTagsCmd(toolsCpDir)

	// test args
	{
		expected := []string{
			javaPath, "-classpath", toolsCpDir,
			"clojure.main",
			"-m", "clojure.tools.deps.alpha.script.resolve-tags",
			"--deps-file=deps.edn",
		}

		if len(cmd.Args) != len(expected) {
			t.Errorf("resolveTagsCmd failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("resolveTagsCmd failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}
}

func TestMakeClassPathCmd(t *testing.T) {
	toolsCpDir := "test-cpTools-dir"

	conf := t4cConfig{
		configUser:    "test-config-user",
		configProject: "test-config-project",
		libsFile:      "test-libs-file",
		cpFile:        "test-cp-file",
		jvmFile:       "test-jvm-file",
		mainFile:      "test-main-file",
		toolsArgs: []string{
			"arg1",
			"arg2",
			"arg3",
			"",
			"",
		},
	}

	cmd := makeClassPathCmd(&conf, toolsCpDir)

	// test args
	{
		expected := []string{
			javaPath, "-classpath", toolsCpDir,
			"clojure.main",
			"-m", "clojure.tools.deps.alpha.script.make-classpath2",
			"--config-user", conf.configUser,
			"--config-project", conf.configProject,
			"--libs-file", conf.libsFile,
			"--cp-file", conf.cpFile,
			"--jvm-file", conf.jvmFile,
			"--main-file", conf.mainFile,
		}
		expected = append(expected, conf.toolsArgs...)
		expected = removeEmpty(expected)

		if len(cmd.Args) != len(expected) {
			t.Errorf("makeClassPathCmd failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("makeClassPathCmd failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}
}

func TestGeneratePomCmd(t *testing.T) {
	toolsCpDir := "test-cpTools-dir"

	conf := t4cConfig{
		configUser:    "test-config-user",
		configProject: "test-config-project",
		toolsArgs: []string{
			"arg1",
			"arg2",
			"arg3",
			"",
			"",
		},
	}

	cmd := generatePomCmd(&conf, toolsCpDir)

	// test args
	{
		expected := []string{
			javaPath, "-classpath", toolsCpDir,
			"clojure.main",
			"-m", "clojure.tools.deps.alpha.script.generate-manifest2",
			"--config-user", conf.configUser,
			"--config-project", conf.configProject,
			"--gen=pom",
		}
		expected = append(expected, conf.toolsArgs...)
		expected = removeEmpty(expected)

		if len(cmd.Args) != len(expected) {
			t.Errorf("generatePomCmd failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("generatePomCmd failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}
}

func TestPrintTreeCmd(t *testing.T) {
	toolsCpDir := "test-cpTools-dir"

	conf := t4cConfig{
		libsFile: "test-libs-file",
	}

	cmd := printTreeCmd(&conf, toolsCpDir)

	// test args
	{
		expected := []string{
			javaPath, "-classpath", toolsCpDir,
			"clojure.main",
			"-m", "clojure.tools.deps.alpha.script.print-tree",
			"--libs-file", conf.libsFile,
		}

		if len(cmd.Args) != len(expected) {
			t.Errorf("printTreeCmd failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("printTreeCmd failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}
}

func TestClojureCmd(t *testing.T) {
	jvmCacheOpts := []string{
		"-jvmCacheOpts",
		"test",
	}
	jvmOpts := []string{
		"-jvmOpts",
		"test",
	}
	conf := t4cConfig{
		libsFile: "test-libs-file",
	}
	cp := "test-class-path"
	mainCacheOpts := []string{
		"-mainCacheOpts",
		"test",
	}
	clojureArgs := []string{
		"-clojureArgs",
		"test",
	}

	cmd := clojureCmd(jvmCacheOpts, jvmOpts, conf.libsFile, cp,
		mainCacheOpts, clojureArgs, false)

	expected := []string{}

	// test clojure args
	{
		expected = []string{javaPath}
		expected = append(expected, jvmCacheOpts...)
		expected = append(expected, jvmOpts...)
		expected = append(expected, "-Dclojure.libfile="+conf.libsFile, "-classpath", cp, "clojure.main")
		expected = append(expected, mainCacheOpts...)
		expected = append(expected, clojureArgs...)
		expected = removeEmpty(expected)

		if len(cmd.Args) != len(expected) {
			t.Errorf("clojureCmd failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("clojureCmd failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}

	cmd = clojureCmd(jvmCacheOpts, jvmOpts, conf.libsFile, cp,
		mainCacheOpts, clojureArgs, true)

	// test clj (rlwrap'ed) args
	{
		if runtime.GOOS != "windows" {
			expected = append(rlwrapArgs(), expected...)
		}

		if len(cmd.Args) != len(expected) {
			t.Errorf("rlwrap'ed clojureCmd failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("rlwrap'ed clojureCmd failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}
}

func TestStart(t *testing.T) {
	cmd := *exec.Command("echo", "test")

	err := start(cmd)
	if err != nil {
		t.Errorf("start cmd failed, error: %v", err)
	}

	cmd = *exec.Command("not-existing-cmd", "test")
	err = start(cmd)
	if err == nil {
		t.Error("non existing cmd started, with no error")
	}
}

func TestSafeStart(t *testing.T) {
	cmd := *exec.Command("echo", "test")

	err := safeStart(cmd)
	if err != nil {
		t.Errorf("safeStart cmd failed, error: %v", err)
	}

	cmd = *exec.Command("not-existing-cmd", "test")
	err = safeStart(cmd)
	if err == nil {
		t.Error("non existing cmd started, with no error")
	}
}
