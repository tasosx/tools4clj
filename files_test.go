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
	"testing"
	"time"
)

func TestGetCljCacheDir(t *testing.T) {
	// cache used environment variables
	// and prepare their deferred restore
	envCljCache, found := os.LookupEnv("CLJ_CACHE")
	if found {
		defer os.Setenv("CLJ_CACHE", envCljCache)
	}
	envXDGCache, found := os.LookupEnv("XDG_CACHE_HOME")
	if found {
		defer os.Setenv("XDG_CACHE_HOME", envXDGCache)
	}

	// --------------------------------------------------

	configDir := "configdir"
	tmpDir := "tmp-dir"

	// CLJ_CACHE is set
	{
		os.Setenv("CLJ_CACHE", tmpDir)
		expected := path.Join(tmpDir)
		res, err := getCljCacheDir(configDir)
		if err != nil {
			t.Error("could not get CLJ_CACHE based dir")
		}
		if res != expected {
			t.Errorf("wrong cljCacjeDir dir, expected `%v`, got `%v`", expected, res)
		}
		os.Unsetenv("CLJ_CACHE")
	}

	// XDG_CACHE_HOME is set
	{
		os.Setenv("XDG_CACHE_HOME", tmpDir)
		expected := path.Join(tmpDir, "clojure")
		res, err := getCljCacheDir(configDir)
		if err != nil {
			t.Error("could not get XDG_CACHE_HOME based dir")
		}
		if res != expected {
			t.Errorf("wrong cljCacjeDir dir, expected `%v`, got `%v`", expected, res)
		}
		os.Unsetenv("XDG_CACHE_HOME")
	}

	// No environment variable is set
	{
		expected := path.Join(configDir, ".cpcache")
		res, err := getCljCacheDir(configDir)
		if err != nil {
			t.Error("could not get configDir based dir")
		}
		if res != expected {
			t.Errorf("wrong cljCacjeDir, expected `%v`, got `%v`", expected, res)
		}
	}
}

func TestGetConfigDir(t *testing.T) {
	// cache used environment variables
	// and prepare their deferred restore
	envCljCache, found := os.LookupEnv("CLJ_CONFIG")
	if found {
		defer os.Setenv("CLJ_CONFIG", envCljCache)
	}
	envXDGHome, found := os.LookupEnv("XDG_CONFIG_HOME")
	if found {
		defer os.Setenv("XDG_CONFIG_HOME", envXDGHome)
	}
	envHome, found := os.LookupEnv("HOME")
	if found {
		defer os.Setenv("HOME", envHome)
	}

	// --------------------------------------------------

	// No specific environment variable is set
	{
		os.Unsetenv("CLJ_CONFIG")
		os.Unsetenv("XDG_CONFIG_HOME")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Errorf("could not get user home dir: %v", err)
			t.Fail()
		}

		expected := path.Join(homeDir, ".clojure")
		res, err := getConfigDir()
		if err != nil {
			t.Error("could not get configDir based dir")
		}
		if res != expected {
			t.Errorf("wrong cljCacjeDir dir, expected `%v`, got `%v`", expected, res)
		}
	}

	tmpDir := "tmp-dir"

	// CLJ_CONFIG is set
	{
		os.Setenv("CLJ_CONFIG", tmpDir)
		expected := path.Join(tmpDir)
		res, err := getConfigDir()
		if err != nil {
			t.Error("could not get CLJ_CONFIG based dir")
		}
		if res != expected {
			t.Errorf("wrong configDir dir, expected `%v`, got `%v`", expected, res)
		}
		os.Unsetenv("CLJ_CONFIG")
	}

	// XDG_CONFIG_HOME is set
	{
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		expected := path.Join(tmpDir, "clojure")
		res, err := getConfigDir()
		if err != nil {
			t.Error("could not get XDG_CONFIG_HOME based dir")
		}
		if res != expected {
			t.Errorf("wrong configDir dir, expected `%v`, got `%v`", expected, res)
		}
		os.Unsetenv("XDG_CONFIG_HOME")
	}

	// HOME is set
	{
		os.Setenv("HOME", tmpDir)
		expected := path.Join(tmpDir, ".clojure")
		res, err := getConfigDir()
		if err != nil {
			t.Error("could not get HOME based dir")
		}
		if res != expected {
			t.Errorf("wrong configDir dir, expected `%v`, got `%v`", expected, res)
		}
		os.Unsetenv("HOME")
	}
}

func TestGetCljToolsDir(t *testing.T) {
	expected := path.Join("test", "tools")
	res := getCljToolsDir("test")

	if res != expected {
		t.Errorf("wrong cljToolsDir, expected `%v`, got `%v`", expected, res)
	}
}

func TestFileExists(t *testing.T) {
	tmpTestFile := "test-filename.txt"
	err := os.WriteFile(tmpTestFile, []byte("Hello"), 0755)
	if err != nil {
		t.Errorf("Unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(tmpTestFile)
	}

	if fileExists(tmpTestFile) == false {
		t.Error("created file does not exist")
	}

	if fileExists("not-existing-"+tmpTestFile) != false {
		t.Error("not existing file exists")
	}
}

func TestDirExists(t *testing.T) {
	tmpTestDir := "test-dir"
	err := os.Mkdir(tmpTestDir, os.ModePerm)
	if err != nil {
		t.Errorf("Unable to create dir: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(tmpTestDir)
	}

	if dirExists(tmpTestDir) == false {
		t.Error("created dir does not exist")
	}

	if dirExists("not-existing-"+tmpTestDir) != false {
		t.Error("not existing dir exists...")
	}
}

func TestIsReadOnlyDir(t *testing.T) {
	tmpTestDir := "test-dir"
	err := os.Mkdir(tmpTestDir, os.ModeDir)
	if err != nil {
		t.Errorf("Unable to create dir: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(tmpTestDir)
	}

	if !isReadOnlyDir("not-existing-" + tmpTestDir) {
		t.Error("not existing dir... is writable!")
	}

	if !isReadOnlyDir(tmpTestDir) {
		t.Error("created read only dir is writable")
	}

	// change mode to read-write
	err = os.Chmod(tmpTestDir, os.ModePerm)
	if err != nil {
		t.Error("could not change mode to RW")
	}

	if isReadOnlyDir(tmpTestDir) {
		t.Error("created dir is not writable")
	}
}

func TestCopyFile(t *testing.T) {
	tmpTestFile1 := "test-filename1.txt"
	tmpTestFile2 := "test-filename2.txt"

	err := os.WriteFile(tmpTestFile1, []byte("Hello"), 0755)
	if err != nil {
		t.Errorf("Unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(tmpTestFile1)
	}

	err = copyFile(tmpTestFile2, tmpTestFile1)
	if err != nil {
		t.Errorf("failed to copy file %v to %v", tmpTestFile1, tmpTestFile2)
	} else {
		defer os.Remove(tmpTestFile2)
	}
}

func TestCheckIsNewerFile(t *testing.T) {
	tmpTestFile1 := "test-filename1.txt"
	tmpTestFile2 := "test-filename2.txt"

	// first file does not exist

	newer, err := checkIsNewerFile(tmpTestFile1, tmpTestFile2)
	if err != nil {
		t.Errorf("error on newer file check: %v", err)
	}
	if newer != false {
		t.Errorf("newer file check, when not existing first file, expecting: %v, got %v", false, newer)
	}

	// --------------------------------------------------

	err = os.WriteFile(tmpTestFile1, []byte("Hello 1"), 0755)
	if err != nil {
		t.Errorf("Unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(tmpTestFile1)
	}

	// same file compared

	newer, err = checkIsNewerFile(tmpTestFile1, tmpTestFile1)
	if err != nil {
		t.Errorf("error on newer file check: %v", err)
	}
	if newer != false {
		t.Errorf("newer file check, same file, expecting: %v, got %v", false, newer)
	}

	// first file exists

	newer, err = checkIsNewerFile(tmpTestFile1, tmpTestFile2)
	if err != nil {
		t.Errorf("error on newer file check: %v", err)
	}
	if newer != true {
		t.Errorf("newer file check, when not existing second file, expecting: %v, got %v", true, newer)
	}

	// wait a bit, and create the second file

	time.Sleep(100 * time.Millisecond)

	err = os.WriteFile(tmpTestFile2, []byte("Hello 2"), 0755)
	if err != nil {
		t.Errorf("Unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(tmpTestFile2)
	}

	// both files exist

	newer, err = checkIsNewerFile(tmpTestFile2, tmpTestFile1)
	if err != nil {
		t.Errorf("error on newer file check: %v", err)
	}
	if newer != true {
		t.Errorf("newer file check, first to second, expecting: %v, got %v", true, newer)
	}

	newer, err = checkIsNewerFile(tmpTestFile1, tmpTestFile2)
	if err != nil {
		t.Errorf("error on newer file check: %v", err)
	}
	if newer != false {
		t.Errorf("newer file check, first to second, expecting: %v, got %v", false, newer)
	}
}

func TestGetExecCpFile(t *testing.T) {
	plainTextCp := path.Join("this", "is", "a", "text", "cp")

	execCp := getExecCpFile(plainTextCp, "exec.jar")
	if execCp != plainTextCp+string(os.PathListSeparator)+"exec.jar" {
		t.Errorf("error on getting exec cp, got: %v", execCp)
	}

	cpFile := "@this_is_a_file.cp"

	execCp = getExecCpFile(cpFile, "exec.jar")
	if execCp != cpFile+".exec" {
		t.Errorf("error on getting exec cp, got: %v", execCp)
	}
}
