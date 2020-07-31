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

Tools for Clojure (t4c) are binary alternatives 
to the official CLI shell scripts: 'clj' and 'clojure'. 

---

You use the Clojure tools ('clj' or 'clojure') to run Clojure programs
on the JVM, e.g. to start a REPL or invoke a specific function with data.
The Clojure tools will configure the JVM process by defining a classpath
(of desired libraries), an execution environment (JVM options) and
specifying a main class and args. 

Using a deps.edn file (or files), you tell Clojure where your source code
resides and what libraries you need. Clojure will then calculate the full
set of required libraries and a classpath, caching expensive parts of this
process for better performance.

The internal steps of the Clojure tools, as well as the Clojure functions
you intend to run, are parameterized by data structures, often maps. Shell
command lines are not optimized for passing nested data, so instead you
will put the data structures in your deps.edn file and refer to them on the
command line via 'aliases' - keywords that name data structures.

'clj' and 'clojure' differ in that 'clj' has extra support for use as a REPL
in a terminal, and should be preferred unless you don't want that support,
then use 'clojure'.

Usage:
  Start a REPL   clj     [t4c-opt*] [clj-opt*] [init-opt*]
  Exec function  clojure [t4c-opt*] [clj-opt*] -X:an-alias [kpath v]*
                 clojure [t4c-opt*] [clj-opt*] -Ffn [kpath v]*
  Run main       clojure [t4c-opt*] [clj-opt*] [--] [init-opt*] [main-opt] [arg*]

The clj-opts are used to build the java-opts and classpath:
  -Jopt          Pass opt through in java_opts, ex: -J-Xmx512m
  -Oalias...     Concatenated jvm option aliases, ex: -O:mem
  -Ralias...     Concatenated resolve-deps aliases, ex: -R:bench:1.9
  -Calias...     Concatenated make-classpath aliases, ex: -C:dev
  -Malias...     Concatenated main option aliases, ex: -M:test
  -Talias...     Concatenated tool option aliases, ex: -T:format-src
  -Aalias...     Concatenated aliases of any kind, ex: -A:dev:mem
  -Xalias K V... Exec alias to invoke a function that takes a map, with keypath/value overrides
  -Fmy/fn K V... Exec function myfn that takes a map, with keypath/value overrides
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
--rebel        Used only by clj tool. Launches clj in a rebel-readline wrapper, 
               instead of the default rlwrap
--native-args  Use unaltered, native, command line args parsing on Windows
               no need to set it on other platforms

For more info, see:
  https://clojure.org/guides/deps_and_cli
  https://clojure.org/reference/repl_and_main
  https://github.com/tasosx/tools4clj
`

const (
	version        = "1.10.1.615"
	depsEDN        = "deps.edn"
	exampleDepsEDN = "example-deps.edn"
	cljExecCLJ     = "clj_exec.clj"
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
	basisFile     string
	toolsArgs     []string
}

func buildCmdConfigs(conf *t4cConfig, cacheDir string, ck string) {
	conf.libsFile = path.Join(cacheDir, ck+".libs")
	conf.cpFile = path.Join(cacheDir, ck+".cp")
	conf.jvmFile = path.Join(cacheDir, ck+".jvm")
	conf.mainFile = path.Join(cacheDir, ck+".main")
	conf.basisFile = path.Join(cacheDir, ck+".basis")
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
		fileExists(path.Join(toolsDir, exampleDepsEDN)) &&
		fileExists(path.Join(toolsDir, cljExecCLJ)) {
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
	err = pickFiles(toolsDir, tarPathTmp, []string{
		depsEDN,
		exampleDepsEDN,
		cljExecCLJ,
		toolsJar,
	})
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
