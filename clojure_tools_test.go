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
	"strings"
	"testing"
)

func TestBuildCmdConfigs(t *testing.T) {
	conf := t4cConfig{
		libsFile: "",
		cpFile:   "",
		jvmFile:  "",
		mainFile: "",
	}

	// --------------------------------------------------

	testCacheDir := "test-cache-dir"
	ck := "00000000"
	buildCmdConfigs(&conf, testCacheDir, ck)

	expected := path.Join(testCacheDir, ck+".libs")
	if conf.libsFile != expected {
		t.Errorf("build cmd configs failed, libsFile expected %v, got %v", expected, conf.libsFile)
	}

	expected = path.Join(testCacheDir, ck+".cp")
	if conf.cpFile != expected {
		t.Errorf("build cmd configs failed, cpFile expected %v, got %v", expected, conf.cpFile)
	}

	expected = path.Join(testCacheDir, ck+".jvm")
	if conf.jvmFile != expected {
		t.Errorf("build cmd configs failed, jvmFile expected %v, got %v", expected, conf.jvmFile)
	}

	expected = path.Join(testCacheDir, ck+".main")
	if conf.mainFile != expected {
		t.Errorf("build cmd configs failed, mainFile expected %v, got %v", expected, conf.mainFile)
	}
}

func TestGetTools4CljPath(t *testing.T) {
	tools4CljDir, err := getTools4CljPath()
	if err != nil {
		t.Errorf("failed to get tools4clj path: %v", err)
	}

	expected := path.Join(t4cHome, version)
	if strings.HasSuffix(tools4CljDir, expected) == false {
		t.Errorf("failed to get tools4clj path, expected ...%v, got %v", expected, tools4CljDir)
	}
}

func TestGetJavaPath(t *testing.T) {
	javaPath, err := getJavaPath()
	if err != nil {
		t.Errorf("failed to get java path: %v", err)
	}
	if fileExists(javaPath) == false {
		t.Errorf("java path: %v does not exist", javaPath)
	}
}

func TestGetOverridenJavaPath(t *testing.T) {
	overridesJava := "overrides_java"
	os.Setenv("JAVA_CMD", overridesJava)

	javaPath, err := getJavaPath()
	if err != nil {
		t.Errorf("failed to get java path: %v", err)
	}
	if javaPath != overridesJava {
		t.Errorf("overriding java path failed: %v", javaPath)
	}
}

func TestGetToolsCp(t *testing.T) {
	dir := ""
	toolsCp, err := getToolsCp(dir)
	if err == nil {
		t.Errorf("expected to get an error for empty tools jar path")
	}

	dir = "test-tools4clj-dir"

	toolsCp, err = getToolsCp(dir)
	if err != nil {
		t.Errorf("failed to get tools class path: %v", err)
	}
	expected := path.Join(dir, toolsJar)
	if toolsCp != expected {
		t.Errorf("failed getting tools class path, expected %v, got %v", expected, toolsCp)
	}
}

func TestGetExecJarPath(t *testing.T) {
	dir := ""
	execCp, err := getExecCp(dir)
	if err == nil {
		t.Errorf("expected to get an error for empty exec jar path")
	}

	dir = "test-tools4clj-dir"

	execCp, err = getExecCp(dir)
	if err != nil {
		t.Errorf("failed to get exec jar path: %v", err)
	}
	expected := path.Join(dir, execJar)
	if execCp != expected {
		t.Errorf("failed getting exec jar path, expected %v, got %v", expected, execCp)
	}
}

func TestGetClojureTools(t *testing.T) {
	dir := ""
	err := getClojureTools(dir)
	if err == nil {
		t.Errorf("expected to get an error for empty download dir creation")
	}

	dir = "testdata"

	err = getClojureTools(dir)
	if err != nil {
		t.Errorf("failed to download clojure tools with error: %v", err)
	}
}

func TestGetConfigPaths(t *testing.T) {
	toolsDir := "testdata"
	dir := "test-config-dir"

	config.configUser = ""
	res := getConfigPaths(&config, dir, toolsDir, false)
	if len(res) != 3 {
		t.Errorf("failed to get all config paths, expected %v, got %v", 3, len(res))
	}
	if len(config.configUser) == 0 {
		t.Error("failed to set `configStr`")
	}

	config.configUser = ""
	res = getConfigPaths(&config, dir, toolsDir, true)
	if len(res) != 2 {
		t.Errorf("failed to get all config paths, expected %v, got %v", 2, len(res))
	}
	if len(config.configUser) != 0 {
		t.Errorf("expected empty `configUser`, got %v", config.configUser)
	}
}

func TestCopyExampleDeps(t *testing.T) {
	toolsDir := "testdata"
	dir := "test-dest-dir"

	os.Mkdir(dir, os.ModePerm)
	defer os.Remove(dir)
	defer os.Remove(path.Join(dir, depsEDN))

	err := copyExampleDeps(dir, toolsDir)
	if err != nil {
		t.Errorf("failed to copy example deps: %v", err)
	}
}
