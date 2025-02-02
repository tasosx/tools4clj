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
	"hash/crc32"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

type allOpts struct {
	Clj        cljOpts
	Init       initOpts
	Main       mainOpts
	Args       []string
	NativeArgs bool
	Rlwrap     bool
	Mode       string
}

type cljOpts struct {
	JvmOpts        []string
	MainAliases    string
	ReplAliases    []string
	ToolAliases    string
	ToolName       string
	ExecAliases    string
	DepsData       string
	PrintClassPath bool
	ForceCP        string
	Prep           bool
	Repro          bool
	Pom            bool
	Tree           bool
	Force          bool
	Verbose        bool
	Describe       bool
	Threads        int
	Trace          bool
	InvalidOption  string
}

type initOpts struct {
	Init   string
	Eval   string
	Report string
}

type mainOpts struct {
	MainArgs []string
	Repl     bool
	Help     bool
	HelpArg  string
}

func read(all *allOpts, args []string, cljRun bool) (bool, error) {
	if len(args) == 0 {
		return false, errors.New("missing application argument (0)")
	}

	var i = 1

	i, err := setT4COpts(all, args, i, cljRun)
	if err != nil {
		return false, err
	}

	// resolve "linuxized" windows command line args
	args, err = linuxize(args, all.NativeArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return true, nil
	}

	i, err = setCljOpts(all, args, i)
	if err != nil {
		return false, err
	} else if i < 0 {
		return true, nil
	}

	i, err = setInitOpts(all, args, i)
	if err != nil {
		return false, err
	}

	i, err = setMainOpts(all, args, i)
	if err != nil {
		return false, err
	}

	if all.Main.Help && (len(all.Clj.MainAliases) > 0 || len(all.Clj.ReplAliases) > 0) {
		all.Main.Help = false
		all.Args = append(all.Args, all.Main.HelpArg)
		return false, nil
	}

	if i < len(args) {
		all.Args = append(all.Args, args[i:]...)
	}

	return false, nil
}

func setT4COpts(all *allOpts, args []string, pos int, cljRun bool) (int, error) {
	all.Rlwrap = cljRun
	all.NativeArgs = (runtime.GOOS != "windows")

	rebel := false
	for {
		if pos >= len(args) {
			break
		}

		if args[pos] == "--rebel" {
			if !cljRun {
				return pos, errors.New("readline option " + args[pos] + " can only be used with clj")
			}
			if rebel {
				return pos, errors.New("readline option " + args[pos] + " defined more than one time")
			}

			all.Rlwrap = false
			all.Clj.DepsData = rebelSdepsArg
			all.Main.MainArgs = append(all.Main.MainArgs, "-m", rebelMainArg)

			rebel = true

		} else if args[pos] == "--native-args" {
			all.NativeArgs = true
		} else {
			// move to the next options group
			break
		}

		pos++
	}
	return pos, nil
}

func setCljOpts(all *allOpts, args []string, pos int) (int, error) {
	all.Mode = "repl"
	for {
		if pos >= len(args) {
			break
		}

		if args[pos] == "-version" {
			fmt.Fprintln(os.Stderr, "Clojure CLI version "+version)
			return -1, nil
		} else if args[pos] == "--version" {
			fmt.Fprintln(os.Stdout, "Clojure CLI version "+version)
			return -1, nil
		} else if strings.HasPrefix(args[pos], "-J") {
			all.Clj.JvmOpts = append(all.Clj.JvmOpts, strings.TrimPrefix(args[pos], "-J"))
		} else if strings.HasPrefix(args[pos], "-R") {
			return pos, errors.New("-R is no longer supported, use -A with repl, -M for main, -X for exec, -T for tool")
		} else if strings.HasPrefix(args[pos], "-C") {
			return pos, errors.New("-C is no longer supported, use -A with repl, -M for main, -X for exec, -T for tool")
		} else if strings.HasPrefix(args[pos], "-O") {
			return pos, errors.New("-O is no longer supported, use -A with repl, -M for main, -X for exec, -T for tool")
		} else if args[pos] == "-A" {
			return pos, errors.New("-A requires an alias")
		} else if strings.HasPrefix(args[pos], "-A") {
			all.Clj.ReplAliases = append(all.Clj.ReplAliases, strings.TrimPrefix(args[pos], "-A"))
		} else if args[pos] == "-M" {
			all.Mode = "main"
			// move to the next options group
			pos++
			break
		} else if strings.HasPrefix(args[pos], "-M") {
			all.Mode = "main"
			all.Clj.MainAliases = strings.TrimPrefix(args[pos], "-M")
			// move to the next options group
			pos++
			break
		} else if args[pos] == "-X" {
			all.Mode = "exec"
			// move to the next options group
			pos++
			break
		} else if strings.HasPrefix(args[pos], "-X") {
			all.Mode = "exec"
			all.Clj.ExecAliases = strings.TrimPrefix(args[pos], "-X")
			// move to the next options group
			pos++
			break
		} else if strings.HasPrefix(args[pos], "-T:") {
			all.Mode = "tool"
			all.Clj.ToolAliases = strings.TrimPrefix(args[pos], "-T")
			pos++
			break
		} else if args[pos] == "-T" {
			all.Mode = "tool"
			pos++
			break
		} else if strings.HasPrefix(args[pos], "-T") {
			all.Mode = "tool"
			all.Clj.ToolName = strings.TrimPrefix(args[pos], "-T")
			pos++
			break
		} else if args[pos] == "-P" {
			all.Clj.Prep = true
		} else if args[pos] == "-Sdeps" {
			if len(all.Clj.DepsData) > 0 {
				return pos, errors.New("deps data option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("deps data value (EDN) not defined for -Sdeps option")
			}
			pos++
			all.Clj.DepsData = args[pos]
		} else if args[pos] == "-Spath" {
			all.Clj.PrintClassPath = true
		} else if args[pos] == "-Scp" {
			if len(all.Clj.ForceCP) > 0 {
				return pos, errors.New("classpath option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("classpath value (CP) not defined for -Scp option")
			}
			pos++
			all.Clj.ForceCP = args[pos]
		} else if args[pos] == "-Srepro" {
			all.Clj.Repro = true
		} else if args[pos] == "-Sforce" {
			all.Clj.Force = true
		} else if args[pos] == "-Spom" {
			all.Clj.Pom = true
		} else if args[pos] == "-Stree" {
			all.Clj.Tree = true
		} else if args[pos] == "-Sresolve-tags" {
			return pos, errors.New("option changed, use: clj -X:deps git-resolve-tags")
		} else if args[pos] == "-Sverbose" {
			all.Clj.Verbose = true
		} else if args[pos] == "-Sdescribe" {
			all.Clj.Describe = true
		} else if args[pos] == "-Sthreads" {
			if all.Clj.Threads > 0 {
				return pos, errors.New("threads option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("threads value (N) not defined for -Sthreads option")
			}
			pos++
			i, err := strconv.Atoi(args[pos])
			if err != nil {
				return pos, errors.New("threads value '" + args[pos] + "' is not a number")
			}
			all.Clj.Threads = i
		} else if args[pos] == "-Strace" {
			all.Clj.Trace = true
		} else if strings.HasPrefix(args[pos], "-S") {
			return pos, errors.New("invalid option:" + args[pos])
		} else if args[pos] == "--" {
			// explicit move to the next options group
			pos++
			break
		} else {
			// move to the next options group
			break
		}

		pos++
	}

	return pos, nil
}

func setInitOpts(all *allOpts, args []string, pos int) (int, error) {
	for {
		if pos >= len(args) {
			break
		}

		if args[pos] == "-i" || args[pos] == "--init" {
			if len(all.Init.Init) > 0 {
				return pos, errors.New("init option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("init path not defined for " + args[pos] + " option")
			}
			pos++
			all.Init.Init = args[pos]
		} else if args[pos] == "-e" || args[pos] == "--eval" {
			if len(all.Init.Eval) > 0 {
				return pos, errors.New("eval option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("eval string not defined for " + args[pos] + " option")
			}
			pos++
			all.Init.Eval = args[pos]
		} else if args[pos] == "--report" {
			if len(all.Init.Report) > 0 {
				return pos, errors.New("report option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("report target not defined for " + args[pos] + " option")
			}
			pos++
			if len(args[pos]) == 0 {
				return pos, errors.New("empty report target is not valid for " + args[pos-1] + " option")
			}
			all.Init.Report = args[pos]
		} else {
			// move to the next options group
			break
		}

		pos++
	}

	return pos, nil
}

func setMainOpts(all *allOpts, args []string, pos int) (int, error) {
	if pos >= len(args) {
		return pos, nil
	}

	if args[pos] == "-m" || args[pos] == "--main" {
		if pos+1 > len(args)-1 {
			return pos, errors.New("main ns-name not defined for " + args[pos] + " option")
		}
		pos++
		all.Main.MainArgs = append(all.Main.MainArgs, "-m", args[pos])
		pos++
	} else if args[pos] == "-r" || args[pos] == "--repl" {
		all.Main.Repl = true
		pos++
	} else if args[pos] == "-h" || args[pos] == "-?" || args[pos] == "--help" {
		all.Main.Help = true
		all.Main.HelpArg = args[pos]
		pos++
	}

	return pos, nil
}

func use(options *allOpts) error {
	// Determine user config directory
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// If user config directory does not exist, create it
	// and copy there the example deps edn file
	err = copyExampleDepsEdn(configDir, tools4CljDir)
	if err != nil {
		return err
	}

	// Determine user clj tools directory
	cljToolsDir := getCljToolsDir(configDir)

	// If clj tools directory does not exist, create it
	// and copy there the tools edn file
	err = copyToolsEdn(cljToolsDir, tools4CljDir)
	if err != nil {
		return err
	}

	// Chain deps.edn in config paths. repro=skip config dir
	config.configProject = "deps.edn"
	configPaths := getConfigPaths(&config, configDir, tools4CljDir, options.Clj.Repro)

	// Determine whether to use user or project cache
	cacheDir := ""
	cacheDirKey := ""
	if fileExists("deps.edn") {
		if !isReadOnlyDir(".") {
			cacheDir = ".cpcache"
		} else {
			cacheDirKey, err = os.Getwd()
			if err != nil {
				return err
			}
			cacheDir, err = getCljCacheDir(configDir)
			if err != nil {
				return err
			}
		}
	} else {
		// Use user cache directory
		cacheDir, err = getCljCacheDir(configDir)
		if err != nil {
			return err
		}
	}

	// Calculate a checksum based on current options and config paths
	ck := checksumOf(options, configPaths, cacheDirKey)

	// Build the file parameters:
	// cpFile, jvmFile, mainFile
	buildCmdConfigs(&config, cacheDir, ck)

	if options.Clj.Verbose {
		fmt.Fprintln(os.Stderr, "version      = "+version)
		fmt.Fprintln(os.Stderr, "install_dir  = "+tools4CljDir)
		fmt.Fprintln(os.Stderr, "config_dir   = "+configDir)
		fmt.Fprintln(os.Stderr, "config_paths = "+join(configPaths, " "))
		fmt.Fprintln(os.Stderr, "cache_dir    = "+cacheDir)
		fmt.Fprintln(os.Stderr, "cp_file      = "+config.cpFile)
	}

	// Check for stale classpath
	stale, err := isStale(options, config, configPaths)
	if err != nil {
		return err
	}

	// Make tools args if needed
	buildToolsArgs(&config, stale, options)

	// If stale, run make-classpath to refresh cached classpath
	if stale && !options.Clj.Describe {
		if options.Clj.Verbose {
			fmt.Fprintln(os.Stderr, "Refreshing classpath")
		}
		err := start(makeClassPathCmd(&config, toolsCp))
		if err != nil {
			return err
		}
	}

	// Get active classpath to use
	cp, err := activeClassPath(options, config)
	if err != nil {
		return err
	}

	// Finally...
	if options.Clj.Pom {
		err := start(generatePomCmd(&config, toolsCp))
		if err != nil {
			return err
		}
	} else if options.Clj.Prep {
		return nil
	} else if options.Clj.PrintClassPath {
		fmt.Println(cp)
	} else if options.Clj.Describe {
		pathVector := ""
		for _, configPath := range configPaths {
			pathVector += "\"" + configPath + "\" "
		}
		fmt.Println(argsDescription(pathVector, tools4CljDir, configDir, cacheDir, &config, options))
	} else if options.Clj.Tree {
		return nil
	} else if options.Clj.Trace {
		fmt.Fprintln(os.Stderr, "Wrote trace.edn")
	} else if options.Mode == "exec" || options.Mode == "tool" {
		jvmCacheOpts, err := getCacheOpts(config.jvmFile)
		if err != nil {
			return err
		}
		err = safeStart(clojureExecuteCmd(jvmCacheOpts, options.Clj.JvmOpts, config.basisFile, execCp, cp, options.Args))
		if err != nil {
			return err
		}
	} else {
		if options.Mode == "repl" && len(options.Args) > 0 {
			fmt.Fprintln(os.Stderr, "WARNING: Implicit use of clojure.main with options is deprecated, use -M $@")
		}
		jvmCacheOpts, err := getCacheOpts(config.jvmFile)
		if err != nil {
			return err
		}
		mainCacheOpts, err := getCacheOpts(config.mainFile)
		if err != nil {
			return err
		}

		clojureArgs := []string{}
		clojureArgs = append(clojureArgs, getInitArgs(options)...)
		clojureArgs = append(clojureArgs, options.Main.MainArgs...)
		clojureArgs = append(clojureArgs, options.Args...)

		err = safeStart(clojureCmd(jvmCacheOpts, options.Clj.JvmOpts,
			config.basisFile, cp, mainCacheOpts, clojureArgs, options.Rlwrap))
		if err != nil {
			return err
		}
	}

	return nil
}

func checksumOf(options *allOpts, configPaths []string, cacheDirKey string) string {
	var cacheVersion = "6"
	prep := join([]string{
		cacheVersion,
		cacheDirKey,
		join(options.Clj.ReplAliases, ""),
		options.Clj.ExecAliases,
		options.Clj.MainAliases,
		options.Clj.DepsData,
		options.Clj.ToolName,
		options.Clj.ToolAliases}, "|")
	for _, v := range configPaths {
		if fileExists(v) {
			prep += "|" + v
		} else {
			prep += "|NIL"
		}
	}
	c := crc32.ChecksumIEEE([]byte(prep))
	return fmt.Sprintf("%d", c)
}

func isStale(options *allOpts, config t4cConfig, configPaths []string) (bool, error) {
	if options.Clj.Force || options.Clj.Trace || options.Clj.Tree || options.Clj.Prep || !fileExists(config.cpFile) {
		return true, nil
	} else {
		newer := false
		if len(options.Clj.ToolName) > 0 {
			configDir, err := getConfigDir()
			if err != nil {
				return false, err
			}
			cljToolsDir := getCljToolsDir(configDir)
			newer, err = checkIsNewerFile(path.Join(cljToolsDir, options.Clj.ToolName+".edn"), config.cpFile)
			if err != nil {
				return false, err
			}
		}
		if newer {
			return true, nil
		} else {
			if fileExists(config.cpFile) {
				b, err := ioutil.ReadFile(config.cpFile)
				if err != nil {
					return false, err
				}
				for _, entry := range strings.Split(string(b), ":") {
					if strings.HasSuffix(entry, ".jar") && !fileExists(entry) {
						return true, nil
					}
				}
			}

			for _, path := range configPaths {
				newer, err := checkIsNewerFile(path, config.cpFile)
				if err != nil {
					return false, err
				}
				if newer {
					return true, nil
				}
			}

			if fileExists(config.manifestFile) {
				manifests, err := readNonEmptyLines(config.manifestFile)
				if err != nil {
					return false, err
				}
				for _, manifest := range manifests {
					if !fileExists(manifest) {
						return true, nil
					}
					newer, err := checkIsNewerFile(manifest, config.cpFile)
					if err != nil {
						return false, err
					}
					if newer {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

func buildToolsArgs(config *t4cConfig, stale bool, options *allOpts) {
	if stale || options.Clj.Pom {
		config.toolsArgs = []string{}
		if len(options.Clj.DepsData) > 0 {
			config.toolsArgs = append(config.toolsArgs, "--config-data", options.Clj.DepsData)
		}
		if len(options.Clj.MainAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-M"+options.Clj.MainAliases)
		}
		if len(options.Clj.ReplAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-A"+join(options.Clj.ReplAliases, ""))
		}
		if len(options.Clj.ExecAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-X"+options.Clj.ExecAliases)
		}
		if options.Mode == "tool" {
			config.toolsArgs = append(config.toolsArgs, "--tool-mode")
		}
		if len(options.Clj.ToolName) > 0 {
			config.toolsArgs = append(config.toolsArgs, "--tool-name", options.Clj.ToolName)
		}
		if len(options.Clj.ToolAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-T"+options.Clj.ToolAliases)
		}
		if len(options.Clj.ForceCP) > 0 {
			config.toolsArgs = append(config.toolsArgs, "--skip-cp")
		}
		if options.Clj.Threads > 0 {
			config.toolsArgs = append(config.toolsArgs, "--threads")
			config.toolsArgs = append(config.toolsArgs, strconv.Itoa(options.Clj.Threads))
		}
		if options.Clj.Tree {
			config.toolsArgs = append(config.toolsArgs, "--tree")
		}
		if options.Clj.Trace {
			config.toolsArgs = append(config.toolsArgs, "--trace")
		}
	}
}

func activeClassPath(options *allOpts, config t4cConfig) (string, error) {
	var cp string
	if options.Clj.Describe {
		cp = ""
	} else if len(options.Clj.ForceCP) > 0 {
		cp = options.Clj.ForceCP
	} else {
		b, err := ioutil.ReadFile(config.cpFile)
		if err != nil {
			return "", err
		}
		if len(b) > 2048 {
			cp = "@" + config.cpFile
		} else {
			cp = string(b)
		}
	}
	return cp, nil
}

func argsDescription(pathVector string, toolsDir string, configDir string, cacheDir string, config *t4cConfig, options *allOpts) string {
	return `{:version "` + version + `"
 :config-files [` + escOnWindows(pathVector) + `]
 :config-user "` + escOnWindows(config.configUser) + `"
 :config-project "` + escOnWindows(config.configProject) + `"
 :install-dir "` + escOnWindows(toolsDir) + `"
 :config-dir "` + escOnWindows(configDir) + `"
 :cache-dir "` + escOnWindows(cacheDir) + `"
 :force ` + strconv.FormatBool(options.Clj.Force) + `
 :repro ` + strconv.FormatBool(options.Clj.Repro) + `
 :main-aliases "` + options.Clj.MainAliases + `"
 :repl-aliases "` + join(options.Clj.ReplAliases, " ") + `"}`
}

func escOnWindows(path string) string {
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, "\\", "\\\\")
	}
	return path
}

func getInitArgs(options *allOpts) []string {
	initArgs := []string{}
	if options.Init.Init != "" {
		initArgs = append(initArgs, []string{`-i`, options.Init.Init}...)
	}
	if options.Init.Eval != "" {
		initArgs = append(initArgs, []string{`-e`, options.Init.Eval}...)
	}
	if options.Init.Report != "" {
		initArgs = append(initArgs, []string{`--report`, options.Init.Report}...)
	}
	return initArgs
}

func getCacheOpts(file string) ([]string, error) {
	cacheOpts := []string{}
	if fileExists(file) {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return []string{}, err
		}
		cacheOpts = strings.Split(string(b), "\n")
	}
	return cacheOpts, nil
}
