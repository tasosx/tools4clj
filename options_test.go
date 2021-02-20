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
	"strings"
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
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // clojure, NativeArgs set
		[]string{"clojure", "--native-args"},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
}

var testT4CRlwrapItems = []TestReadItem{
	{ // clj, Rlwrap
		[]string{"clj"},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     true,
			Mode:       "repl",
		},
		"",
	},
	{ // clj, rebel-readline
		[]string{"clj", "--rebel"},
		allOpts{
			Clj: cljOpts{
				DepsData: rebelSdepsArg,
			},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", rebelMainArg},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // clj, rebel-readline defined twice
		[]string{"clj", "--rebel", "--rebel"},
		allOpts{
			Clj: cljOpts{
				DepsData: rebelSdepsArg,
			},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", rebelMainArg},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"readline option --rebel defined more than one time",
	},
	{ // clojure, can not use rebel-readline
		[]string{"clojure", "--rebel"},
		allOpts{
			Clj: cljOpts{
				DepsData: rebelSdepsArg,
			},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", rebelMainArg},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"readline option --rebel can only be used with clj",
	},
}

var testMainItems = []TestReadItem{
	{ // abnormal totally missing args
		[]string{},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"missing application argument (0)",
	},
	{ // no args
		[]string{"clojure"},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // help arg, long
		[]string{"clojure", "--help"},
		allOpts{
			Clj:  cljOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Help:    true,
				HelpArg: "--help",
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // help arg, short -h
		[]string{"clojure", "-h"},
		allOpts{
			Clj:  cljOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Help:    true,
				HelpArg: "-h",
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // help arg, short -?
		[]string{"clojure", "-?"},
		allOpts{
			Clj:  cljOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Help:    true,
				HelpArg: "-?",
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // main arg, missing namespace
		[]string{"clojure",
			"-m",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"main ns-name not defined for -m option",
	},
	{ // main arg, short -m
		[]string{"clojure",
			"-m", "namespace_name",
		},
		allOpts{
			Clj:  cljOpts{},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", "namespace_name"},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // main arg, long --main
		[]string{"clojure",
			"--main", "namespace_name",
		},
		allOpts{
			Clj:  cljOpts{},
			Init: initOpts{},
			Main: mainOpts{
				MainArgs: []string{"-m", "namespace_name"},
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // repl arg, short -r
		[]string{"clojure", "-r"},
		allOpts{
			Clj:  cljOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Repl: true,
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // repl arg, long --repl
		[]string{"clojure", "--repl"},
		allOpts{
			Clj:  cljOpts{},
			Init: initOpts{},
			Main: mainOpts{
				Repl: true,
			},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // path arg
		[]string{"clojure",
			"this_is_a_path",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{"this_is_a_path"},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // standard input arg -
		[]string{"clojure",
			"-",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{"-"},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // path arg with extra args
		[]string{"clojure",
			"this_is_a_path",
			"--extra1", "-e2",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{"this_is_a_path", "--extra1", "-e2"},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
}

var testDepItems = []TestReadItem{
	{ // not valid Dep -S option
		[]string{"clojure", "-S"},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"invalid option:-S",
	},
	{ // not valid Dep -S option
		[]string{"clojure", "-Sany"},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"invalid option:-Sany",
	},
	{ // more Dep options
		[]string{"clojure",
			"-JargJ",
			"-R:argR",
			"-C:argC",
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
			"-Sverbose",
			"-Sdescribe",
			"-Sthreads",
			"42",
			"-Strace",
		},
		allOpts{
			Clj: cljOpts{
				JvmOpts:          []string{"argJ"},
				ResolveAliases:   ":argR",
				ClassPathAliases: ":argC",
				ReplAliases:      ":argA",
				DepsData:         `{:deps {clansi {:mvn/version "1.0.0"}}}`,
				PrintClassPath:   true,
				ForceCP:          `src:target:/classpath`,
				Repro:            true,
				Force:            true,
				Pom:              true,
				Tree:             true,
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
			Mode:       "repl",
		},
		"",
	},
	{ // multiple Dep options
		[]string{"clojure",
			"-JargJ1", "-JargJ2",
			"-R:argR1", "-R:argR2",
			"-C:argC1", "-C:argC2",
			"-A:argA1", "-A:argA2",
			"-M:argM1",
		},
		allOpts{
			Clj: cljOpts{
				JvmOpts:          []string{"argJ1", "argJ2"},
				ResolveAliases:   ":argR1:argR2",
				ClassPathAliases: ":argC1:argC2",
				ReplAliases:      ":argA1:argA2",
				MainAliases:      ":argM1",
			},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "main",
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
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
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
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
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
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"threads option -Sthreads defined more than one time",
	},
	{ // missing value for Dep option: -Sdeps
		[]string{"clojure",
			"-Sdeps",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"deps data value (EDN) not defined for -Sdeps option",
	},
	{ // missing value for Dep option: -Scp
		[]string{"clojure",
			"-Scp",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"classpath value (CP) not defined for -Scp option",
	},
	{ // missing value for Dep option: -Sthreads
		[]string{"clojure",
			"-Sthreads",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"threads value (N) not defined for -Sthreads option",
	},
	{ // not a number value for Dep option: -Sthreads
		[]string{"clojure",
			"-Sthreads",
			"not-a-number",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"threads value '" + "not-a-number" + "' is not a number",
	},
	{ // pass remaining dep options to clojure.main, by using Dep option: --
		[]string{"clojure",
			"-A:argA1",
			"--",
			"-A:argA2",
		},
		allOpts{
			Clj: cljOpts{
				ReplAliases: ":argA1",
			},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{"-A:argA2"},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // dep main (no alliases)
		[]string{"clojure",
			"-M",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "main",
		},
		"",
	},
	{ // dep main alliases along with help option
		[]string{"clojure",
			"-M:argM",
			"-h",
		},
		allOpts{
			Clj: cljOpts{
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
			Mode:       "main",
		},
		"",
	},
	{ // dep all alliases along with help option
		[]string{"clojure",
			"-A:argA",
			"--help",
		},
		allOpts{
			Clj: cljOpts{
				ReplAliases: ":argA",
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
			Mode:       "repl",
		},
		"",
	},
	{ // clojure exec -X (no alias)
		[]string{"clojure",
			"-X",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "exec",
		},
		"",
	},
	{ // clojure exec -X (alias)
		[]string{"clojure",
			"-X:foo",
		},
		allOpts{
			Clj: cljOpts{
				ExecAliases: ":foo",
			},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "exec",
		},
		"",
	},
	{ // clojure prep -P
		[]string{"clojure",
			"-P",
		},
		allOpts{
			Clj: cljOpts{
				Prep: true,
			},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // not supported: -O
		[]string{"clojure",
			"-O:argO",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"-O is no longer supported, use -A with repl, -M for main, or -X for exec",
	},
	{ // not supported: -T
		[]string{"clojure",
			"-T:argT",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"-T is no longer supported, use -A with repl, -M for main, or -X for exec",
	},
	{ // not supported: -Sresolve-tags
		[]string{"clojure",
			"-Sresolve-tags",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"Option changed, use: clj -X:deps git-resolve-tags",
	},
	{ // not supported: -A without an alias
		[]string{"clojure",
			"-A",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"-A requires an alias",
	},
}

var testInitItems = []TestReadItem{
	{ // missing init path
		[]string{"clojure",
			"-i",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"init path not defined for -i option",
	},
	{ // init path, short -i
		[]string{"clojure",
			"-i", "init_path_file",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Init: "init_path_file",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // init path, long --init
		[]string{"clojure",
			"--init", "init_path_file",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Init: "init_path_file",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // duplicate init path
		[]string{"clojure",
			"-i", "init_path_file",
			"--init", "other_init_path_file",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Init: "init_path_file",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"init option --init defined more than one time",
	},
	{ // missing eval string
		[]string{"clojure",
			"-e",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"eval string not defined for -e option",
	},
	{ // eval string, short -e
		[]string{"clojure",
			"-e", "eval_string",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Eval: "eval_string",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // eval string, long --eval
		[]string{"clojure",
			"--eval", "eval_string",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Eval: "eval_string",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // duplicate eval string
		[]string{"clojure",
			"-e", "eval_string",
			"--eval", "other_eval_string",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Eval: "eval_string",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"eval option --eval defined more than one time",
	},
	{ // missing report target
		[]string{"clojure",
			"--report",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"report target not defined for --report option",
	},
	{ // invalid empty report target
		[]string{"clojure",
			"--report", "",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"empty report target is not valid for --report option",
	},
	{ // accepted report target: file
		[]string{"clojure",
			"--report", "test-filename.txt",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Report: "test-filename.txt",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // accepted report target: stderr
		[]string{"clojure",
			"--report", "stderr",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Report: "stderr",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // accepted report target: none
		[]string{"clojure",
			"--report", "none",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Report: "none",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"",
	},
	{ // duplicate report targets
		[]string{"clojure",
			"--report", "stderr",
			"--report", "none",
		},
		allOpts{
			Clj: cljOpts{},
			Init: initOpts{
				Report: "stderr",
			},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"report option --report defined more than one time",
	},
}

var testVersionPrintItems = []TestReadItem{
	{ // print version on stderr and exit
		[]string{"clojure",
			"-version",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"expected to exit",
	},
	{ // print version on stderr and exit
		[]string{"clojure",
			"--version",
		},
		allOpts{
			Clj:        cljOpts{},
			Init:       initOpts{},
			Main:       mainOpts{},
			Args:       []string{},
			NativeArgs: true,
			Rlwrap:     false,
			Mode:       "repl",
		},
		"expected to exit",
	},
}

func TestRead(t *testing.T) {
	testItems := []TestReadItem{}
	testItems = append(testItems, testT4CNativeArgsItems...)
	testItems = append(testItems, testT4CRlwrapItems...)
	testItems = append(testItems, testMainItems...)
	testItems = append(testItems, testDepItems...)
	testItems = append(testItems, testInitItems...)
	testItems = append(testItems, testVersionPrintItems...)

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
		exit, err := read(&opts, v.inputArgs, clj)
		if err != nil {
			if v.errorExpected == "" || v.errorExpected != err.Error() {
				t.Errorf("could not read args %v, error: %v", v.inputArgs, err)
			}
		} else if exit {
			if v.errorExpected != "expected to exit" {
				t.Errorf("could not exit after: %v", v.inputArgs)
			}
		} else {
			readOpts := fmt.Sprintf("%+v", opts)
			expectedOpts := fmt.Sprintf("%+v", v.expected)
			if readOpts != expectedOpts {
				t.Errorf("read options failed, \nexpected\n%v, \ngot \n%v", expectedOpts, readOpts)
			}
		}
	}
}

func TestChecksumOf(t *testing.T) {
	// input
	options := allOpts{
		Clj: cljOpts{
			ReplAliases: ":argA",
			MainAliases: ":argM",
			DepsData:    `{:deps {clansi {:mvn/version "1.0.0"}}}`,
		},
	}
	// not existing config paths
	configPaths := []string{
		"filepath1.edn",
		"filepath2.edn",
	}
	// output
	expected := "2414352071"

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
	expected = "414039739"

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
	options.Clj.Force = true
	options.Clj.Trace = false
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
	options.Clj.Force = false
	options.Clj.Trace = true
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
	options.Clj.Force = false
	options.Clj.Trace = false
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
	options.Clj.Force = false
	options.Clj.Trace = false
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
	options.Clj.Force = false
	options.Clj.Trace = false
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
	options.Clj.Force = false
	options.Clj.Trace = false
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
	options.Clj.Force = false
	options.Clj.Trace = false
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
		Clj: cljOpts{
			DepsData:         `{:deps {clansi {:mvn/version "1.0.0"}}}`,
			ResolveAliases:   ":argR",
			ClassPathAliases: ":argC",
			MainAliases:      ":argM",
			ReplAliases:      ":argA",
			ExecAliases:      ":argX",
			ForceCP:          "forced-class-path",
			Pom:              false,
			Threads:          42,
			Trace:            true,
		},
	}
	config := t4cConfig{}
	stale := false

	// toolsArgs not changed when: Dep.Pom == false, stale == false
	options.Clj.Pom = false
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
	options.Clj.Pom = true
	config.toolsArgs = []string{"not_changed"}
	stale = false
	expected = []string{
		"--config-data",
		`{:deps {clansi {:mvn/version "1.0.0"}}}`,
		"-R:argR",
		"-C:argC",
		"-M:argM",
		"-A:argA",
		"-X:argX",
		"--skip-cp",
		"--threads",
		"42",
		"--trace",
	}

	buildToolsArgs(&config, stale, &options)
	args = fmt.Sprintf("%+v", config.toolsArgs)
	expectedArgs = fmt.Sprintf("%+v", expected)
	if len(config.toolsArgs) != len(expected) || args != expectedArgs {
		t.Errorf("buildToolsArgs failed, \nexpected %v, \ngot %v", expected, config.toolsArgs)
	}

	// toolsArgs changed when: Dep.Pom == false, stale == true
	options.Clj.Pom = false
	config.toolsArgs = []string{"not_changed"}
	stale = true
	expected = []string{
		"--config-data",
		`{:deps {clansi {:mvn/version "1.0.0"}}}`,
		"-R:argR",
		"-C:argC",
		"-M:argM",
		"-A:argA",
		"-X:argX",
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

	// toolsArgs changed when: Clj.Tree == true, stale == true
	options.Clj.Tree = true
	config.toolsArgs = []string{"not_changed"}
	stale = true
	expected = []string{
		"--config-data",
		`{:deps {clansi {:mvn/version "1.0.0"}}}`,
		"-R:argR",
		"-C:argC",
		"-M:argM",
		"-A:argA",
		"-X:argX",
		"--skip-cp",
		"--threads",
		"42",
		"--tree",
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

	// files to use
	cpFile := "cpfile.cp"
	cpFileContent := "Hello"
	cpFileLarge := "cpfile_large.cp"
	cpFileLargeContent := strings.Repeat("Clojure", 1+2048/len("Clojure"))

	// create cp file
	err := ioutil.WriteFile(cpFile, []byte(cpFileContent), 0755)
	if err != nil {
		t.Errorf("unable to write file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(cpFile)
	}

	// create large cp file (>2048 bytes)
	err = ioutil.WriteFile(cpFileLarge, []byte(cpFileLargeContent), 0755)
	if err != nil {
		t.Errorf("unable to write large file: %v", err)
		t.FailNow()
	} else {
		defer os.Remove(cpFileLarge)
	}

	// no options defined, no class path file
	res, err := activeClassPath(&options, config)
	if err == nil {
		t.Error("expecting file open error")
	}

	// no options defined, not existing class path file
	config.cpFile = "notexisting_cpfile.cp"
	res, err = activeClassPath(&options, config)
	if err == nil {
		t.Error("expecting file open error")
	}

	// no options defined, existing class path file
	config.cpFile = cpFile
	expected := cpFileContent

	res, err = activeClassPath(&options, config)
	if err != nil {
		t.Errorf("activeClassPath failed, with error: %v", err)
	}
	if res != expected {
		t.Errorf("activeClassPath failed, expected %v, got %v", "", res)
	}

	// no options defined, existing class path large file
	config.cpFile = cpFileLarge
	expected = "@" + cpFileLarge

	res, err = activeClassPath(&options, config)
	if err != nil {
		t.Errorf("activeClassPath failed, with error: %v", err)
	}
	if res != expected {
		t.Errorf("activeClassPath failed, expected %v, got %v", "", res)
	}

	// Clj.Describe == true and Clj.ForceCP is set, existing class path file
	options.Clj.Describe = true
	options.Clj.ForceCP = "forced_cp"
	config.cpFile = cpFile
	expected = ""

	res, err = activeClassPath(&options, config)
	if err != nil {
		t.Errorf("activeClassPath failed, with error: %v", err)
	}
	if res != expected {
		t.Errorf("activeClassPath failed, expected %v, got %v", "", res)
	}

	// Clj.Describe == false and Clj.ForceCP is set, existing class path file
	options.Clj.Describe = false
	options.Clj.ForceCP = "forced_cp"
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
		Clj: cljOpts{
			Force:            true,
			Repro:            true,
			ResolveAliases:   ":argR",
			ClassPathAliases: ":argC",
			MainAliases:      ":argM",
			ReplAliases:      ":argA",
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
 :force ` + strconv.FormatBool(options.Clj.Force) + `
 :repro ` + strconv.FormatBool(options.Clj.Repro) + `
 :main-aliases "` + options.Clj.MainAliases + `"
 :repl-aliases "` + options.Clj.ReplAliases + `"}`

	res := argsDescription(pathVector, toolsDir, configDir, cacheDir, &config, &options)
	if res != expected {
		t.Errorf("argsDescription failed, \nexpected\n%v, \ngot \n%v", expected, res)
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
