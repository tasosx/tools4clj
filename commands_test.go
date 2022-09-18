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
	"os/exec"
	"runtime"
	"testing"
)

func TestMakeClassPathCmd(t *testing.T) {
	toolsCpDir := "test-cpTools-dir"

	conf := t4cConfig{
		configUser:    "test-config-user",
		configProject: "test-config-project",
		libsFile:      "test-libs-file",
		basisFile:     "test-basis-file",
		cpFile:        "test-cp-file",
		jvmFile:       "test-jvm-file",
		mainFile:      "test-main-file",
		manifestFile:  "test-manifest-file",
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
			"--basis-file", conf.basisFile,
			"--cp-file", conf.cpFile,
			"--jvm-file", conf.jvmFile,
			"--main-file", conf.mainFile,
			"--manifest-file", conf.manifestFile,
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

	// test when CLJ_JVM_OPTS environment variable is set
	os.Setenv("CLJ_JVM_OPTS", "CLJ_JVM_OPTS_VALUE")

	cmd = makeClassPathCmd(&conf, toolsCpDir)

	os.Unsetenv("CLJ_JVM_OPTS")

	// test args
	{
		expected := []string{
			javaPath, "CLJ_JVM_OPTS_VALUE", "-classpath", toolsCpDir,
			"clojure.main",
			"-m", "clojure.tools.deps.alpha.script.make-classpath2",
			"--config-user", conf.configUser,
			"--config-project", conf.configProject,
			"--libs-file", conf.libsFile,
			"--basis-file", conf.basisFile,
			"--cp-file", conf.cpFile,
			"--jvm-file", conf.jvmFile,
			"--main-file", conf.mainFile,
			"--manifest-file", conf.manifestFile,
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

			for i1, v1 := range cmd.Args {
				t.Errorf("expected %v = %v", i1, v1)
			}

			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("generatePomCmd failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}

	// test when CLJ_JVM_OPTS environment variable is set
	os.Setenv("CLJ_JVM_OPTS", "CLJ_JVM_OPTS_VALUE")

	cmd = generatePomCmd(&conf, toolsCpDir)

	os.Unsetenv("CLJ_JVM_OPTS")

	// test args
	{
		expected := []string{
			javaPath, "CLJ_JVM_OPTS_VALUE", "-classpath", toolsCpDir,
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

func TestClojureExecuteCmd(t *testing.T) {
	jvmCacheOpts := []string{
		"-jvmCacheOpts",
		"test",
	}
	jvmOpts := []string{
		"-jvmOpts",
		"test",
	}
	conf := t4cConfig{
		basisFile: "test-basis-file",
	}
	cp := "test-class-path"
	execJarPath := "exec-jar-path"

	// test clojure -X execute, no aliases, no args
	{
		args := []string{}

		cmd := clojureExecuteCmd(jvmCacheOpts, jvmOpts, conf.basisFile,
			execJarPath, cp, args)

		expected := []string{javaPath}

		expected = append(expected, jvmCacheOpts...)
		expected = append(expected, jvmOpts...)
		expected = append(expected, "-Dclojure.basis="+conf.basisFile,
			"-classpath", cp+string(os.PathListSeparator)+execJarPath)
		expected = append(expected, "clojure.main", "-m", "clojure.run.exec")

		expected = removeEmpty(expected)

		if len(cmd.Args) != len(expected) {
			t.Errorf("clojureExecuteCmd (-X) failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("clojureExecuteCmd (-X) failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}

	// test clojure -X execute, with aliases, no args
	{
		args := []string{"-X:foo"}

		cmd := clojureExecuteCmd(jvmCacheOpts, jvmOpts, conf.basisFile,
			execJarPath, cp, args)

		expected := []string{javaPath}

		expected = append(expected, jvmCacheOpts...)
		expected = append(expected, jvmOpts...)
		expected = append(expected, "-Dclojure.basis="+conf.basisFile,
			"-classpath", cp+string(os.PathListSeparator)+execJarPath)
		expected = append(expected, "clojure.main", "-m", "clojure.run.exec")
		expected = append(expected, args...)

		expected = removeEmpty(expected)

		if len(cmd.Args) != len(expected) {
			t.Errorf("clojureExecuteCmd (-X:foo) failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("clojureExecuteCmd (-X:foo) failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}

	// test clojure -X execute, no aliases, with args
	{
		args := []string{"arg1", "arg2"}

		cmd := clojureExecuteCmd(jvmCacheOpts, jvmOpts, conf.basisFile,
			execJarPath, cp, args)

		expected := []string{javaPath}

		expected = append(expected, jvmCacheOpts...)
		expected = append(expected, jvmOpts...)
		expected = append(expected, "-Dclojure.basis="+conf.basisFile,
			"-classpath", cp+string(os.PathListSeparator)+execJarPath)
		expected = append(expected, "clojure.main", "-m", "clojure.run.exec")
		expected = append(expected, args...)

		expected = removeEmpty(expected)

		if len(cmd.Args) != len(expected) {
			t.Errorf("clojureExecuteCmd (-X arg1 arg2) failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("clojureExecuteCmd (-X arg1 arg2) failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}

	// test clojure -X execute, with aliases, and args
	{
		args := []string{"-X:foo", "arg1", "arg2"}

		cmd := clojureExecuteCmd(jvmCacheOpts, jvmOpts, conf.basisFile,
			execJarPath, cp, args)

		expected := []string{javaPath}

		expected = append(expected, jvmCacheOpts...)
		expected = append(expected, jvmOpts...)
		expected = append(expected, "-Dclojure.basis="+conf.basisFile,
			"-classpath", cp+string(os.PathListSeparator)+execJarPath)
		expected = append(expected, "clojure.main", "-m", "clojure.run.exec")
		expected = append(expected, args...)

		expected = removeEmpty(expected)

		if len(cmd.Args) != len(expected) {
			t.Errorf("clojureExecuteCmd (-X:foo arg1 arg2) failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("clojureExecuteCmd (-X:foo arg1 arg2) failed, arg expected %v, got %v", expected[i], v)
			}
		}
	}

	// test clojure -X execute, with aliases, and args, when JAVA_OPTS environment variable is set
	{
		args := []string{"-X:foo", "arg1", "arg2"}

		os.Setenv("JAVA_OPTS", "JAVA_OPTS_VALUE")

		cmd := clojureExecuteCmd(jvmCacheOpts, jvmOpts, conf.basisFile,
			execJarPath, cp, args)

		os.Unsetenv("JAVA_OPTS")

		expected := []string{javaPath}

		expected = append(expected, "JAVA_OPTS_VALUE")
		expected = append(expected, jvmCacheOpts...)
		expected = append(expected, jvmOpts...)
		expected = append(expected, "-Dclojure.basis="+conf.basisFile,
			"-classpath", cp+string(os.PathListSeparator)+execJarPath)
		expected = append(expected, "clojure.main", "-m", "clojure.run.exec")
		expected = append(expected, args...)

		expected = removeEmpty(expected)

		if len(cmd.Args) != len(expected) {
			t.Errorf("clojureExecuteCmd (-X:foo arg1 arg2) failed, args expected %v, got %v", len(expected), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expected[i] {
				t.Errorf("clojureExecuteCmd (-X:foo arg1 arg2) failed, arg expected %v, got %v", expected[i], v)
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

	expected := []string{javaPath}

	// test clojure args
	cmd := clojureCmd(jvmCacheOpts, jvmOpts, conf.libsFile, conf.basisFile,
		cp, mainCacheOpts, clojureArgs, false)

	{
		expected = append(expected, jvmCacheOpts...)
		expected = append(expected, jvmOpts...)
		expected = append(expected, "-Dclojure.libfile="+conf.libsFile, "-Dclojure.basisfile="+conf.basisFile, "-classpath", cp, "clojure.main")
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

	// test clj (rlwrap'ed) args
	cmd = clojureCmd(jvmCacheOpts, jvmOpts, conf.libsFile, conf.basisFile,
		cp, mainCacheOpts, clojureArgs, true)

	{
		expectedWrapped := []string{javaPath}

		if runtime.GOOS != "windows" {
			expectedWrapped = append(rlwrapArgs(), expected...)
		}

		if len(cmd.Args) != len(expectedWrapped) {
			t.Errorf("rlwrap'ed clojureCmd failed, args expected %v, got %v", len(expectedWrapped), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expectedWrapped[i] {
				t.Errorf("rlwrap'ed clojureCmd failed, arg expected %v, got %v", expectedWrapped[i], v)
			}
		}
	}

	// test clojure args, when JAVA_OPTS environment variable is set
	os.Setenv("JAVA_OPTS", "JAVA_OPTS_VALUE")

	cmd = clojureCmd(jvmCacheOpts, jvmOpts, conf.libsFile, conf.basisFile,
		cp, mainCacheOpts, clojureArgs, false)

	os.Unsetenv("JAVA_OPTS")

	{
		expected = []string{javaPath}
		expected = append(expected, "JAVA_OPTS_VALUE")
		expected = append(expected, jvmCacheOpts...)
		expected = append(expected, jvmOpts...)
		expected = append(expected, "-Dclojure.libfile="+conf.libsFile, "-Dclojure.basisfile="+conf.basisFile, "-classpath", cp, "clojure.main")
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

	// test clj (rlwrap'ed) args, when JAVA_OPTS environment variable is set
	os.Setenv("JAVA_OPTS", "JAVA_OPTS_VALUE")

	cmd = clojureCmd(jvmCacheOpts, jvmOpts, conf.libsFile, conf.basisFile,
		cp, mainCacheOpts, clojureArgs, true)

	os.Unsetenv("JAVA_OPTS")

	{
		expectedWrapped := []string{javaPath}

		if runtime.GOOS != "windows" {
			expectedWrapped = append(rlwrapArgs(), expected...)
		}

		if len(cmd.Args) != len(expectedWrapped) {
			t.Errorf("rlwrap'ed clojureCmd failed, args expected %v, got %v", len(expectedWrapped), len(cmd.Args))
			t.FailNow()
		}

		for i, v := range cmd.Args {
			if v != expectedWrapped[i] {
				t.Errorf("rlwrap'ed clojureCmd failed, arg expected %v, got %v", expectedWrapped[i], v)
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
