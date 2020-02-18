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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
)

const usage = `Version: ` + version + ` of clojure tools

Usage:  clojure [t4c-opt*] [dep-opt*] [--] [init-opt*] [main-opt] [arg*]
        clj     [t4c-opt*] [dep-opt*] [--] [init-opt*] [main-opt] [arg*]

clojure executable is a runner for Clojure. clj is its wrapper
for interactive repl use. These executables ultimately 
construct and invoke a command-line of the form:

java [java-opt*] -cp classpath clojure.main [init-opt*] [main-opt] [arg*]

The dep-opts are used to build the java-opts and classpath:
  -Jopt          Pass opt through in java_opts, ex: -J-Xmx512m
  -Oalias...     Concatenated jvm option aliases, ex: -O:mem
  -Ralias...     Concatenated resolve-deps aliases, ex: -R:bench:1.9
  -Calias...     Concatenated make-classpath aliases, ex: -C:dev
  -Malias...     Concatenated main option aliases, ex: -M:test
  -Aalias...     Concatenated aliases of any kind, ex: -A:dev:mem
  -Sdeps EDN     Deps data to use as the last deps file to be merged
  -Spath         Compute classpath and echo to stdout only
  -Scp CP        Do NOT compute or cache classpath, use this one instead
  -Srepro        Ignore the ~/.clojure/deps.edn config file
  -Sforce        Force recomputation of the classpath (don't use the cache)
  -Spom          Generate (or update existing) pom.xml with deps and paths
  -Stree         Print dependency tree
  -Sresolve-tags Resolve git coordinate tags to shas and update deps.edn
  -Sverbose      Print important path info to console
  -Sdescribe     Print environment and command parsing info as data
  -Sthreads N    Set specific number of download threads
  -Strace        Write a trace.edn file that traces deps expansion
  --             Stop parsing dep options and pass remaining arguments to clojure.main

init-opt:
  -i, --init path      Load a file or resource
  -e, --eval string    Eval exprs in string; print non-nil values
      --report target  Report uncaught exception to "file" (default), "stderr", or "none", 
                       overrides System property clojure.main.report

main-opt:
  -m, --main ns-name   Call the -main function from namespace w/args
  -r, --repl           Run a repl
  path                 Run a script from a file or resource
  -                    Run a script from standard input
  -h, -?, --help       Print this help message and exit

t4c-opt:
--rebel        Used only by clj. Launches clj in a rebel-readline wrapper, 
               instead of the rlwrap
--native-args  Use unaltered, native, command line args parsing on Windows
               no need to set it on other platforms

For more info, see:
  https://clojure.org/guides/deps_and_cli
  https://clojure.org/reference/repl_and_main
  https://github.com/tasosx/tools4clj
`

const (
	version        = "1.10.1.510"
	depsEDN        = "deps.edn"
	exampleDepsEDN = "example-deps.edn"
	toolsTarGz     = "clojure-tools-" + version + ".tar.gz"
	toolsURL       = "https://download.clojure.org/install/" + toolsTarGz
	toolsJar       = "clojure-tools-" + version + ".jar"
	t4cHome        = ".tools4clj"
)

var (
	tools4CljDir = ""
	toolsCp      = ""
	javaPath     = ""
	config       t4cConfig
)

type t4cConfig struct {
	configUser    string
	configProject string
	libsFile      string
	cpFile        string
	jvmFile       string
	mainFile      string
	toolsArgs     []string
}

func buildCmdConfigs(conf *t4cConfig, cacheDir string, ck string) {
	conf.libsFile = path.Join(cacheDir, ck+".libs")
	conf.cpFile = path.Join(cacheDir, ck+".cp")
	conf.jvmFile = path.Join(cacheDir, ck+".jvm")
	conf.mainFile = path.Join(cacheDir, ck+".main")
}

func getTools4CljPath() (string, error) {
	env, found := os.LookupEnv("HOME")
	if found {
		return path.Join(env, t4cHome, version), nil
	}
	env, error := os.UserHomeDir()
	if error != nil {
		return "", error
	}
	return path.Join(env, t4cHome, version), nil
}

func getJavaPath() (string, error) {
	var p = exec.Command("java").Path
	if p == "" {
		env, found := os.LookupEnv("JAVA_HOME")
		if !found {
			return "", errors.New("could not find java executable - please set JAVA_HOME")
		}
		p = path.Join(env, "bin", "java")

		if runtime.GOOS == "windows" {
			p += ".exe"
		}
	}
	return p, nil
}

func getToolsCp(toolsDir string) (string, error) {
	if toolsDir == "" {
		return "", errors.New("empty install dir")
	}

	return path.Join(toolsDir, toolsJar), nil
}

func getClojureTools(toolsDir string) error {
	err := os.MkdirAll(toolsDir, os.ModePerm)
	if err != nil {
		return err
	}

	if fileExists(path.Join(toolsDir, toolsJar)) &&
		fileExists(path.Join(toolsDir, depsEDN)) &&
		fileExists(path.Join(toolsDir, exampleDepsEDN)) {
		return nil
	}

	fmt.Println("[t4c] - downloading official clojure tools")

	// download the official clojure tools tar.gr
	var tarPathTmp = path.Join(toolsDir, toolsTarGz)
	if !fileExists(tarPathTmp) {
		err = downloadFile(tarPathTmp, toolsURL)
	}
	if err != nil {
		return err
	}

	fmt.Println("[t4c] - extracting needed clojure tools files")

	// extract the needed files
	err = pickFiles(toolsDir, tarPathTmp, []string{depsEDN, exampleDepsEDN, toolsJar})
	if err != nil {
		return err
	}

	fmt.Println("[t4c] - cleaning up")

	// remove the official clojure tools tar.gz
	err = os.Remove(tarPathTmp)
	if err != nil {
		return err
	}

	return nil
}

func getConfigPaths(conf *t4cConfig, configDir string, toolsDir string, repro bool) []string {
	configPaths := []string{}
	configUser := ""
	if repro {
		configPaths = []string{path.Join(toolsDir, "deps.edn"), "deps.edn"}
	} else {
		configUser = path.Join(configDir, "deps.edn")
		configPaths = []string{path.Join(toolsDir, "deps.edn"), path.Join(configDir, "deps.edn"), "deps.edn"}
	}
	conf.configUser = configUser

	return configPaths
}

func copyExampleDeps(destDir string, toolsDir string) error {
	var t4CDeps = path.Join(toolsDir, exampleDepsEDN)
	var localDeps = path.Join(destDir, depsEDN)

	if fileExists(localDeps) {
		return nil
	}
	if destDir != "" {
		err := os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return copyFile(localDeps, t4CDeps)
}
