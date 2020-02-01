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
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"testing"
	"time"
)

type TestReadItem struct {
	inputArgs     []string
	expected      allOpts
	errorExpected string
}

var testT4CNativeArgsItems = []TestReadItem{
	{ // clojure, NativeArgs not set
		[]string{"clojure"},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // clojure, NativeArgs set
		[]string{"clojure", "--native-args"},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
}

var testT4CRlwrapItems = []TestReadItem{
	{ // clj, Rlwrap
		[]string{"clj"},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     true,
		},
		"",
	},
	{ // clj, rebel-readline
		[]string{"clj", "--rebel"},
		allOpts{
			Dep: depOpts{
				DepsData: rebelSdepsArg,
			},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", rebelMainArg},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // clj, rebel-readline defined twice
		[]string{"clj", "--rebel", "--rebel"},
		allOpts{
			Dep: depOpts{
				DepsData: rebelSdepsArg,
			},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", rebelMainArg},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"readline option --rebel defined more than one time",
	},
	{ // clojure, can not use rebel-readline
		[]string{"clojure", "--rebel"},
		allOpts{
			Dep: depOpts{
				DepsData: rebelSdepsArg,
			},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", rebelMainArg},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"readline option --rebel can only be used with clj",
	},
}

var testMainItems = []TestReadItem{
	{ // abnormal totally missing args
		[]string{},
		allOpts{Dep: depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false},
		"missing application argument (0)",
	},
	{ // no args
		[]string{"clojure"},
		allOpts{Dep: depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // help arg, long
		[]string{"clojure", "--help"},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Help:    true,
				HelpArg: "--help",
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // help arg, short -h
		[]string{"clojure", "-h"},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Help:    true,
				HelpArg: "-h",
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // help arg, short -?
		[]string{"clojure", "-?"},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Help:    true,
				HelpArg: "-?",
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // main arg, missing namespace
		[]string{"clojure",
			"-m",
		},
		allOpts{Dep: depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"main ns-name not defined for -m option",
	},
	{ // main arg, short -m
		[]string{"clojure",
			"-m", "namespace_name",
		},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", "namespace_name"},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // main arg, long --main
		[]string{"clojure",
			"--main", "namespace_name",
		},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", "namespace_name"},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // repl arg, short -r
		[]string{"clojure", "-r"},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Repl: true,
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // repl arg, long --repl
		[]string{"clojure", "--repl"},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Repl: true,
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // path arg
		[]string{"clojure",
			"this_is_a_path",
		},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				PathArg: "this_is_a_path",
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // standard input arg -
		[]string{"clojure",
			"-",
		},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				PathArg: "-",
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // path arg with extra args
		[]string{"clojure",
			"this_is_a_path",
			"--extra1", "-e2",
		},
		allOpts{Dep: depOpts{},
			Init: initOpts{},
			Main: mainOpts{
				PathArg: "this_is_a_path",
			},
			Args:       []string{"--extra1", "-e2"},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
}

var testDepItems = []TestReadItem{
	{ // not valid Dep -S option
		[]string{"clojure", "-S"},
		allOpts{Dep: depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"invalid option:-S",
	},
	{ // not valid Dep -S option
		[]string{"clojure", "-Sany"},
		allOpts{Dep: depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"invalid option:-Sany",
	},
	{ // all Dep options
		[]string{"clojure",
			"-JargJ",
			"-O:argO",
			"-R:argR",
			"-C:argC",
			"-M:argM",
			"-A:argA",
			"-Sdeps",
			`{:deps {clansi {:mvn/version "1.0.0"}}}`,
			"-Spath",
			"-Scp",
			`src:target:/classpath`,
			"-Srepro",
			"-Sforce",
			"-Spom",
			"-Stree",
			"-Sresolve-tags",
			"-Sverbose",
			"-Sdescribe",
			"-Sthreads",
			"42",
			"-Strace",
		},
		allOpts{
			Dep: depOpts{
				JvmOpts:          []string{"argJ"},
				JvmAliases:       ":argO",
				ResolveAliases:   ":argR",
				ClassPathAliases: ":argC",
				MainAliases:      ":argM",
				AllAliases:       ":argA",
				DepsData:         `{:deps {clansi {:mvn/version "1.0.0"}}}`,
				PrintClassPath:   true,
				ForceCP:          `src:target:/classpath`,
				Repro:            true,
				Force:            true,
				Pom:              true,
				Tree:             true,
				ResolveTags:      true,
				Verbose:          true,
				Describe:         true,
				Threads:          42,
				Trace:            true,
			},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // multiple Dep options
		[]string{"clojure",
			"-JargJ1", "-JargJ2",
			"-O:argO1", "-O:argO2",
			"-R:argR1", "-R:argR2",
			"-C:argC1", "-C:argC2",
			"-M:argM1", "-M:argM2",
			"-A:argA1", "-A:argA2",
		},
		allOpts{
			Dep: depOpts{
				JvmOpts:          []string{"argJ1", "argJ2"},
				JvmAliases:       ":argO1:argO2",
				ResolveAliases:   ":argR1:argR2",
				ClassPathAliases: ":argC1:argC2",
				MainAliases:      ":argM1:argM2",
				AllAliases:       ":argA1:argA2",
			},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // not valid Dep multiple option: -Sdeps
		[]string{"clojure",
			"-Sdeps",
			`{:deps {clansi {:mvn/version "1.0.0"}}}`,
			"-Sdeps",
			`{:deps {clansi {:mvn/version "1.0.1"}}}`,
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"deps data option -Sdeps defined more than one time",
	},
	{ // not valid Dep multiple option: -Scp
		[]string{"clojure",
			"-Scp",
			`src:target:/classpath1`,
			"-Scp",
			`src:target:/classpath2`,
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"classpath option -Scp defined more than one time",
	},
	{ // not valid Dep multiple option: -Sthreads
		[]string{"clojure",
			"-Sthreads",
			`1`,
			"-Sthreads",
			`5`,
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"threads option -Sthreads defined more than one time",
	},
	{ // missing value for Dep option: -Sdeps
		[]string{"clojure",
			"-Sdeps",
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"deps data value (EDN) not defined for -Sdeps option",
	},
	{ // missing value for Dep option: -Scp
		[]string{"clojure",
			"-Scp",
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"classpath value (CP) not defined for -Scp option",
	},
	{ // missing value for Dep option: -Sthreads
		[]string{"clojure",
			"-Sthreads",
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"threads value (N) not defined for -Sthreads option",
	},
	{ // not a number value for Dep option: -Sthreads
		[]string{"clojure",
			"-Sthreads",
			"not-a-number",
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"threads value '" + "not-a-number" + "' is not a number",
	},
	{ // dep main alliases along with help option
		[]string{"clojure",
			"-M:argM",
			"-h",
		},
		allOpts{
			Dep: depOpts{
				MainAliases: ":argM",
			},
			Init: initOpts{},
			Main: mainOpts{
				Help:    false,
				HelpArg: "-h",
			},
			Args: []string{
				"-h",
			},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"classpath value (CP) not defined for -Scp option",
	},
	{ // dep all alliases along with help option
		[]string{"clojure",
			"-A:argA",
			"--help",
		},
		allOpts{
			Dep: depOpts{
				AllAliases: ":argA",
			},
			Init: initOpts{},
			Main: mainOpts{
				Help:    false,
				HelpArg: "--help",
			},
			Args: []string{
				"--help",
			},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"classpath value (CP) not defined for -Scp option",
	},
}

var testInitItems = []TestReadItem{
	{ // missing init path
		[]string{"clojure",
			"-i",
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"init path not defined for -i option",
	},
	{ // init path, short -i
		[]string{"clojure",
			"-i", "init_path_file",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Init: "init_path_file",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // init path, long --init
		[]string{"clojure",
			"--init", "init_path_file",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Init: "init_path_file",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // duplicate init path
		[]string{"clojure",
			"-i", "init_path_file",
			"--init", "other_init_path_file",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Init: "init_path_file",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"init option --init defined more than one time",
	},
	{ // missing eval string
		[]string{"clojure",
			"-e",
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"eval string not defined for -e option",
	},
	{ // eval string, short -e
		[]string{"clojure",
			"-e", "eval_string",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Eval: "eval_string",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // eval string, long --eval
		[]string{"clojure",
			"--eval", "eval_string",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Eval: "eval_string",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // duplicate eval string
		[]string{"clojure",
			"-e", "eval_string",
			"--eval", "other_eval_string",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Eval: "eval_string",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"eval option --eval defined more than one time",
	},
	{ // missing report target
		[]string{"clojure",
			"--report",
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"report target not defined for --report option",
	},
	{ // invalid empty report target
		[]string{"clojure",
			"--report", "",
		},
		allOpts{
			Dep:        depOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"empty report target is not valid for --report option",
	},
	{ // accepted report target: file
		[]string{"clojure",
			"--report", "test-filename.txt",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Report: "test-filename.txt",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // accepted report target: stderr
		[]string{"clojure",
			"--report", "stderr",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Report: "stderr",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // accepted report target: none
		[]string{"clojure",
			"--report", "none",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Report: "none",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"",
	},
	{ // duplicate report targets
		[]string{"clojure",
			"--report", "stderr",
			"--report", "none",
		},
		allOpts{
			Dep: depOpts{},
			Init: initOpts{
				Report: "stderr",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
		},
		"report option --report defined more than one time",
	},
}

func TestRead(t *testing.T) {
	testItems := []TestReadItem{}
	testItems = append(testItems, testT4CNativeArgsItems...)
	testItems = append(testItems, testT4CRlwrapItems...)
	testItems = append(testItems, testMainItems...)
	testItems = append(testItems, testDepItems...)
	testItems = append(testItems, testInitItems...)

	for _, v := range testItems {
		// on windows run only native args read test
		if runtime.GOOS == "windows" {
			// inject native args argument in command line input
			if len(v.inputArgs) > 1 {
				v.inputArgs = append([]string{v.inputArgs[0], "--native-args"}, v.inputArgs[1:]...)
			} else if len(v.inputArgs) == 1 {
				v.inputArgs = []string{v.inputArgs[0], "--native-args"}
			}
		}

		opts := allOpts{}
		clj := false
		if len(v.inputArgs) > 0 && v.inputArgs[0] == "clj" {
			clj = true
		}
		err := read(&opts, v.inputArgs, clj)
		if err != nil {
			if v.errorExpected == "" || v.errorExpected != err.Error() {
				t.Errorf("could not read args %v, error: %v", v.inputArgs, err)
			}
		} else {
			readOpts := fmt.Sprintf("%+v", opts)
			expectedOpts := fmt.Sprintf("%+v", v.expected)
			if readOpts != expectedOpts {
				t.Errorf("read options failed, expected %v, got %v", expectedOpts, readOpts)
			}
		}
	}
}

func TestChecksumOf(t *testing.T) {
	// input
	options := allOpts{
		Dep: depOpts{
			ResolveAliases:   ":argR",
			ClassPathAliases: ":argC",
			AllAliases:       ":argA",
			JvmAliases:       ":argJ",
			MainAliases:      ":argM",
			DepsData:         `{:deps {clansi {:mvn/version "1.0.0"}}}`,
		},
	}
	// not existing config paths
	configPaths := []string{
		"filepath1.edn",
		"filepath2.edn",
	}
	// output
	expected := "456246304"

	res := checksumOf(&options, configPaths)
	if res != expected {
		t.Errorf("checksumOf failed, expected %v, got %v", expected, res)
	}

	// create one of the defined config files
	tmpExistingFile := "filepath1.edn"
	err := ioutil.WriteFile(tmpExistingFile, []byte("Hello"), 0755)
	if err != nil {
		t.Errorf("unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(tmpExistingFile)
	}

	// different output is expected
	expected = "4122476052"

	res = checksumOf(&options, configPaths)
	if res != expected {
		t.Errorf("checksumOf failed, expected %v, got %v", expected, res)
	}
}

func TestIsStale(t *testing.T) {
	options := allOpts{}
	config := t4cConfig{}
	configPaths := []string{}

	// files to use
	olderFile := "older_filepath.edn"
	cpFile := "classpathFile.edn"
	newerFile := "newer_filepath.edn"

	// create an, older than cp, config path file
	err := ioutil.WriteFile(olderFile, []byte("Hello"), 0755)
	if err != nil {
		t.Errorf("unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(olderFile)
	}

	time.Sleep(100 * time.Millisecond)

	// create cp file
	err = ioutil.WriteFile(cpFile, []byte("Hello"), 0755)
	if err != nil {
		t.Errorf("unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(cpFile)
	}

	time.Sleep(100 * time.Millisecond)

	// create a, newer than cp, config path file
	err = ioutil.WriteFile(newerFile, []byte("Hello"), 0755)
	if err != nil {
		t.Errorf("unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(newerFile)
	}

	// forced cp
	// inputs
	options.Dep.Force = true
	options.Dep.Trace = false
	config.cpFile = cpFile
	configPaths = []string{newerFile}
	// output
	expected := true

	res, err := isStale(&options, config, configPaths)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		t.FailNow()
	}
	if res != expected {
		t.Errorf("isStale failed, expected %v, got %v", expected, res)
	}

	// trace
	// inputs
	options.Dep.Force = false
	options.Dep.Trace = true
	config.cpFile = cpFile
	configPaths = []string{newerFile}
	// output
	expected = true

	res, err = isStale(&options, config, configPaths)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		t.FailNow()
	}
	if res != expected {
		t.Errorf("isStale failed, expected %v, got %v", expected, res)
	}

	// not existing cpFile
	// inputs
	options.Dep.Force = false
	options.Dep.Trace = false
	config.cpFile = "notexisting_cpFile.edn"
	configPaths = []string{newerFile}
	// output
	expected = true

	res, err = isStale(&options, config, configPaths)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		t.FailNow()
	}
	if res != expected {
		t.Errorf("isStale failed, expected %v, got %v", expected, res)
	}

	// existing cpFile, not existing config paths file
	// inputs
	options.Dep.Force = false
	options.Dep.Trace = false
	config.cpFile = cpFile
	configPaths = []string{"notexisting_file.edn"}
	// output
	expected = false

	res, err = isStale(&options, config, configPaths)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		t.FailNow()
	}
	if res != expected {
		t.Errorf("isStale failed, expected %v, got %v", expected, res)
	}

	// existing cpFile, existing older config paths file
	// inputs
	options.Dep.Force = false
	options.Dep.Trace = false
	config.cpFile = cpFile
	configPaths = []string{olderFile}
	// output
	expected = false

	res, err = isStale(&options, config, configPaths)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		t.FailNow()
	}
	if res != expected {
		t.Errorf("isStale failed, expected %v, got %v", expected, res)
	}

	// existing cpFile, existing newer config paths file
	// inputs
	options.Dep.Force = false
	options.Dep.Trace = false
	config.cpFile = cpFile
	configPaths = []string{newerFile}
	// output
	expected = true

	res, err = isStale(&options, config, configPaths)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		t.FailNow()
	}
	if res != expected {
		t.Errorf("isStale failed, expected %v, got %v", expected, res)
	}

	// existing cpFile, existing at least one newer config paths file
	// inputs
	options.Dep.Force = false
	options.Dep.Trace = false
	config.cpFile = cpFile
	configPaths = []string{olderFile, newerFile}
	// output
	expected = true

	res, err = isStale(&options, config, configPaths)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		t.FailNow()
	}
	if res != expected {
		t.Errorf("isStale failed, expected %v, got %v", expected, res)
	}
}

func TestBuildToolsArgs(t *testing.T) {
	options := allOpts{
		Dep: depOpts{
			DepsData:         `{:deps {clansi {:mvn/version "1.0.0"}}}`,
			ResolveAliases:   ":argR",
			ClassPathAliases: ":argC",
			JvmAliases:       ":argJ",
			MainAliases:      ":argM",
			AllAliases:       ":argA",
			ForceCP:          "forced-class-path",
			Pom:              false,
			Threads:          42,
			Trace:            true,
		},
	}
	config := t4cConfig{}
	stale := false

	// toolsArgs not changed when: Dep.Pom == false, stale == false
	options.Dep.Pom = false
	config.toolsArgs = []string{"not_changed"}
	stale = false
	expected := []string{"not_changed"}

	buildToolsArgs(&config, stale, &options)
	args := fmt.Sprintf("%+v", config.toolsArgs)
	expectedArgs := fmt.Sprintf("%+v", expected)
	if len(config.toolsArgs) != len(expected) || args != expectedArgs {
		t.Errorf("buildToolsArgs failed, expected %v, got %v", expected, config.toolsArgs)
	}

	// toolsArgs changed when: Dep.Pom == true, stale == false
	options.Dep.Pom = true
	config.toolsArgs = []string{"not_changed"}
	stale = false
	expected = []string{
		"--config-data",
		`{:deps {clansi {:mvn/version "1.0.0"}}}`,
		"-R:argR",
		"-C:argC",
		"-J:argJ",
		"-M:argM",
		"-A:argA",
		"--skip-cp",
		"--threads",
		"42",
		"--trace",
	}

	buildToolsArgs(&config, stale, &options)
	args = fmt.Sprintf("%+v", config.toolsArgs)
	expectedArgs = fmt.Sprintf("%+v", expected)
	if len(config.toolsArgs) != len(expected) || args != expectedArgs {
		t.Errorf("buildToolsArgs failed, expected %v, got %v", expected, config.toolsArgs)
	}

	// toolsArgs changed when: Dep.Pom == false, stale == true
	options.Dep.Pom = false
	config.toolsArgs = []string{"not_changed"}
	stale = true
	expected = []string{
		"--config-data",
		`{:deps {clansi {:mvn/version "1.0.0"}}}`,
		"-R:argR",
		"-C:argC",
		"-J:argJ",
		"-M:argM",
		"-A:argA",
		"--skip-cp",
		"--threads",
		"42",
		"--trace",
	}

	buildToolsArgs(&config, stale, &options)
	args = fmt.Sprintf("%+v", config.toolsArgs)
	expectedArgs = fmt.Sprintf("%+v", expected)
	if len(config.toolsArgs) != len(expected) || args != expectedArgs {
		t.Errorf("buildToolsArgs failed, expected %v, got %v", expected, config.toolsArgs)
	}
}

func TestActiveClassPath(t *testing.T) {
	options := allOpts{}
	config := t4cConfig{}

	// file to use
	cpFile := "classpathFile.edn"

	// create cp file
	err := ioutil.WriteFile(cpFile, []byte("Hello"), 0755)
	if err != nil {
		t.Errorf("unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(cpFile)
	}

	// no options defined, no class path file
	res, err := activeClassPath(&options, config)
	if err == nil {
		t.Error("expecting file open error")
	}

	// no options defined, not existing class path file
	config.cpFile = "notexisting_cpfile.edn"
	res, err = activeClassPath(&options, config)
	if err == nil {
		t.Error("expecting file open error")
	}

	// no options defined, existing class path file
	config.cpFile = cpFile
	expected := "Hello"

	res, err = activeClassPath(&options, config)
	if err != nil {
		t.Errorf("activeClassPath failed, with error: %v", err)
	}
	if res != expected {
		t.Errorf("activeClassPath failed, expected %v, got %v", "", res)
	}

	// Dep.Describe == true, existing class path file
	options.Dep.Describe = true
	options.Dep.ForceCP = "forced_cp"
	config.cpFile = cpFile
	expected = ""

	res, err = activeClassPath(&options, config)
	if err != nil {
		t.Errorf("activeClassPath failed, with error: %v", err)
	}
	if res != expected {
		t.Errorf("activeClassPath failed, expected %v, got %v", "", res)
	}

	// Dep.ForceCP == false and Dep.ForceCP is set, existing class path file
	options.Dep.Describe = false
	options.Dep.ForceCP = "forced_cp"
	config.cpFile = cpFile
	expected = "forced_cp"

	res, err = activeClassPath(&options, config)
	if err != nil {
		t.Errorf("activeClassPath failed, with error: %v", err)
	}
	if res != expected {
		t.Errorf("activeClassPath failed, expected %v, got %v", "", res)
	}
}

func TestArgsDescription(t *testing.T) {
	// input
	pathVector := "testPath"
	toolsDir := "testToolsDir"
	configDir := "testConfigDir"
	cacheDir := "testCacheDir"
	config := t4cConfig{
		configUser:    "testConfigUser",
		configProject: "testConfigProject",
	}
	options := allOpts{
		Dep: depOpts{
			Force:            true,
			Repro:            true,
			ResolveAliases:   ":argR",
			ClassPathAliases: ":argC",
			JvmAliases:       ":argJ",
			MainAliases:      ":argM",
			AllAliases:       ":argA",
		},
	}
	// output
	expected := `{:version "` + version + `"
 :config-files [` + pathVector + `]
 :config-user "` + config.configUser + `"
 :config-project "` + config.configProject + `"
 :install-dir "` + toolsDir + `"
 :config-dir "` + configDir + `"
 :cache-dir "` + cacheDir + `"
 :force ` + strconv.FormatBool(options.Dep.Force) + `
 :repro ` + strconv.FormatBool(options.Dep.Repro) + `
 :resolve-aliases "` + options.Dep.ResolveAliases + `"
 :classpath-aliases "` + options.Dep.ClassPathAliases + `"
 :jvm-aliases "` + options.Dep.JvmAliases + `"
 :main-aliases "` + options.Dep.MainAliases + `"
 :all-aliases "` + options.Dep.AllAliases + `"}`

	res := argsDescription(pathVector, toolsDir, configDir, cacheDir, &config, &options)
	if res != expected {
		t.Errorf("argsDescription failed, expected %v, got %v", expected, res)
	}
}

func TestGetInitArgs(t *testing.T) {
	options := allOpts{
		Init: initOpts{
			Init:   "initArg",
			Eval:   "evalArg",
			Report: "reportArg",
		},
	}
	expected := []string{
		`-i`, options.Init.Init,
		`-e`, options.Init.Eval,
		`--report`, options.Init.Report,
	}
	res := getInitArgs(&options)

	args := fmt.Sprintf("%+v", res)
	expectedArgs := fmt.Sprintf("%+v", expected)
	if len(res) != len(expected) || args != expectedArgs {
		t.Errorf("getInitArgs failed, expected %v, got %v", expected, res)
	}
}
