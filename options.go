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
}

type cljOpts struct {
	JvmOpts          []string
	JvmAliases       string
	ResolveAliases   string
	ClassPathAliases string
	MainAliases      string
	ToolAliases      string
	AllAliases       string
	ExecAlias        []string
	DepsData         string
	PrintClassPath   bool
	ForceCP          string
	Repro            bool
	Force            bool
	Pom              bool
	Tree             bool
	ResolveTags      bool
	Verbose          bool
	Describe         bool
	Threads          int
	Trace            bool
	InvalidOption    string
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

func read(all *allOpts, args []string, cljRun bool) error {
	if len(args) == 0 {
		return errors.New("missing application argument (0)")
	}

	var i = 1

	i, err := setT4COpts(all, args, i, cljRun)
	if err != nil {
		return err
	}

	// resolve "linuxized" windows command line args
	args, err = linuxize(args, all.NativeArgs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	i, err = setCljOpts(all, args, i)
	if err != nil {
		return err
	}

	i, err = setInitOpts(all, args, i)
	if err != nil {
		return err
	}

	i, err = setMainOpts(all, args, i)
	if err != nil {
		return err
	}

	if all.Main.Help && (len(all.Clj.MainAliases) > 0 || len(all.Clj.AllAliases) > 0) {
		all.Main.Help = false
		all.Args = append(all.Args, all.Main.HelpArg)
		return nil
	}

	if i < len(args) {
		all.Args = append(all.Args, args[i:]...)
	}

	return nil
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
			if cljRun == false {
				return pos, errors.New("readline option " + args[pos] + " can only be used with clj")
			}
			if rebel == true {
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
	for {
		if pos >= len(args) {
			break
		}

		if strings.HasPrefix(args[pos], "-J") {
			all.Clj.JvmOpts = append(all.Clj.JvmOpts, strings.TrimPrefix(args[pos], "-J"))
		} else if strings.HasPrefix(args[pos], "-O") {
			all.Clj.JvmAliases += strings.TrimPrefix(args[pos], "-O")
		} else if strings.HasPrefix(args[pos], "-R") {
			all.Clj.ResolveAliases += strings.TrimPrefix(args[pos], "-R")
		} else if strings.HasPrefix(args[pos], "-C") {
			all.Clj.ClassPathAliases += strings.TrimPrefix(args[pos], "-C")
		} else if strings.HasPrefix(args[pos], "-M") {
			all.Clj.MainAliases += strings.TrimPrefix(args[pos], "-M")
		} else if strings.HasPrefix(args[pos], "-T") {
			all.Clj.ToolAliases += strings.TrimPrefix(args[pos], "-T")
		} else if strings.HasPrefix(args[pos], "-A") {
			all.Clj.AllAliases += strings.TrimPrefix(args[pos], "-A")
		} else if strings.HasPrefix(args[pos], "-X") || strings.HasPrefix(args[pos], "-F") {
			all.Clj.ExecAlias = append(all.Clj.ExecAlias, args[pos])
			for pos < len(args)-1 && !strings.HasPrefix(args[pos+1], "-") {
				pos++
				all.Clj.ExecAlias = append(all.Clj.ExecAlias, args[pos])
			}
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
			all.Clj.ResolveTags = true
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
	for {
		if pos >= len(args) {
			break
		}

		if args[pos] == "-m" || args[pos] == "--main" {
			if pos+1 > len(args)-1 {
				return pos, errors.New("main ns-name not defined for " + args[pos] + " option")
			}
			pos++
			all.Main.MainArgs = append(all.Main.MainArgs, "-m", args[pos])
		} else if args[pos] == "-r" || args[pos] == "--repl" {
			all.Main.Repl = true
		} else if args[pos] == "-h" || args[pos] == "-?" || args[pos] == "--help" {
			all.Main.Help = true
			all.Main.HelpArg = args[pos]
		} else {
			// move to the next options group
			break
		}
		pos++
		break
	}

	return pos, nil
}

func use(options *allOpts) error {
	// Execute resolve-tags command
	if options.Clj.ResolveTags {
		if fileExists("deps.edn") == false {
			return errors.New("deps.edn does not exist")
		}
		return start(resolveTagsCmd(toolsCp))
	}

	// Determine user config directory
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// If user config directory does not exist, create it
	// and copy there the example deps edn
	err = copyExampleDeps(configDir, tools4CljDir)
	if err != nil {
		return err
	}

	// Chain deps.edn in config paths. repro=skip config dir
	config.configProject = "deps.edn"
	configPaths := getConfigPaths(&config, configDir, tools4CljDir, options.Clj.Repro)

	// Determine whether to use user or project cache
	cacheDir := ""
	if fileExists("deps.edn") {
		cacheDir = ".cpcache"
	} else {
		// Determine user cache directory
		userCacheDir, err := getCljCacheDir(configDir)
		if err != nil {
			return err
		}
		cacheDir = userCacheDir
	}

	// Calculate a checksum based on current options and config paths
	ck := checksumOf(options, configPaths)

	// Build the file parameters:
	// libsFile, cpFile, jvmFile, mainFile
	buildCmdConfigs(&config, cacheDir, ck)

	if options.Clj.Verbose {
		fmt.Println("version      = " + version)
		fmt.Println("install_dir  = " + tools4CljDir)
		fmt.Println("config_dir   = " + configDir)
		fmt.Println("config_paths = " + join(configPaths, " "))
		fmt.Println("cache_dir    = " + cacheDir)
		fmt.Println("cp_file      = " + config.cpFile)
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
			fmt.Println("Refreshing classpath")
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
	} else if options.Clj.PrintClassPath {
		fmt.Println(cp)
	} else if options.Clj.Describe {
		pathVector := ""
		for _, configPath := range configPaths {
			pathVector += "\"" + configPath + "\" "
		}
		fmt.Println(argsDescription(pathVector, tools4CljDir, configDir, cacheDir, &config, options))
	} else if options.Clj.Tree {
		err := start(printTreeCmd(&config, toolsCp))
		if err != nil {
			return err
		}
	} else if options.Clj.Trace {
		fmt.Println("Writing trace.edn")
	} else if len(options.Clj.ExecAlias) > 0 {
		jvmCacheOpts := []string{}
		if fileExists(config.jvmFile) {
			b, err := ioutil.ReadFile(config.jvmFile)
			if err != nil {
				return err
			}
			jvmCacheOpts = strings.Split(string(b), " ")
		}
		err := safeStart(clojureExecuteCmd(jvmCacheOpts, options.Clj.JvmOpts,
			config.basisFile, tools4CljDir, cp, options.Clj.ExecAlias))
		if err != nil {
			return err
		}
	} else {
		jvmCacheOpts := []string{}
		if fileExists(config.jvmFile) {
			b, err := ioutil.ReadFile(config.jvmFile)
			if err != nil {
				return err
			}
			jvmCacheOpts = strings.Split(string(b), " ")
		}
		mainCacheOpts := []string{}
		if fileExists(config.mainFile) {
			b, err := ioutil.ReadFile(config.mainFile)
			if err != nil {
				return err
			}
			mainCacheOpts = strings.Split(string(b), " ")
		}

		clojureArgs := []string{}
		clojureArgs = append(clojureArgs, getInitArgs(options)...)
		clojureArgs = append(clojureArgs, options.Main.MainArgs...)
		clojureArgs = append(clojureArgs, options.Args...)

		err := safeStart(clojureCmd(jvmCacheOpts, options.Clj.JvmOpts,
			config.libsFile, config.basisFile,
			cp, mainCacheOpts, clojureArgs, options.Rlwrap))
		if err != nil {
			return err
		}
	}

	return nil
}

func checksumOf(options *allOpts, configPaths []string) string {
	prep := join([]string{
		options.Clj.ResolveAliases,
		options.Clj.ClassPathAliases,
		options.Clj.AllAliases,
		options.Clj.JvmAliases,
		options.Clj.MainAliases,
		options.Clj.ToolAliases,
		options.Clj.DepsData}, "|")
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
	stale := false
	if options.Clj.Force || options.Clj.Trace || !fileExists(config.cpFile) {
		stale = true
	} else {
		for _, path := range configPaths {
			newer, err := checkIsNewerFile(path, config.cpFile)
			if err != nil {
				return false, err
			}
			if newer {
				stale = true
				break
			}
		}
	}
	return stale, nil
}

func buildToolsArgs(config *t4cConfig, stale bool, options *allOpts) {
	if stale || options.Clj.Pom {
		config.toolsArgs = []string{}
		if len(options.Clj.DepsData) > 0 {
			config.toolsArgs = append(config.toolsArgs, "--config-data", options.Clj.DepsData)
		}
		if len(options.Clj.ResolveAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-R"+options.Clj.ResolveAliases)
		}
		if len(options.Clj.ClassPathAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-C"+options.Clj.ClassPathAliases)
		}
		if len(options.Clj.JvmAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-J"+options.Clj.JvmAliases)
		}
		if len(options.Clj.MainAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-M"+options.Clj.MainAliases)
		}
		if len(options.Clj.ToolAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-T"+options.Clj.ToolAliases)
		}
		if len(options.Clj.AllAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-A"+options.Clj.AllAliases)
		}
		if len(options.Clj.ForceCP) > 0 {
			config.toolsArgs = append(config.toolsArgs, "--skip-cp")
		}
		if options.Clj.Threads > 0 {
			config.toolsArgs = append(config.toolsArgs, "--threads")
			config.toolsArgs = append(config.toolsArgs, strconv.Itoa(options.Clj.Threads))
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
		cp = string(b)
	}
	return cp, nil
}

func argsDescription(pathVector string, toolsDir string, configDir string, cacheDir string, config *t4cConfig, options *allOpts) string {
	return `{:version "` + version + `"
 :config-files [` + pathVector + `]
 :config-user "` + config.configUser + `"
 :config-project "` + config.configProject + `"
 :install-dir "` + toolsDir + `"
 :config-dir "` + configDir + `"
 :cache-dir "` + cacheDir + `"
 :force ` + strconv.FormatBool(options.Clj.Force) + `
 :repro ` + strconv.FormatBool(options.Clj.Repro) + `
 :resolve-aliases "` + options.Clj.ResolveAliases + `"
 :classpath-aliases "` + options.Clj.ClassPathAliases + `"
 :jvm-aliases "` + options.Clj.JvmAliases + `"
 :main-aliases "` + options.Clj.MainAliases + `"
 :tool-aliases "` + options.Clj.ToolAliases + `"
 :all-aliases "` + options.Clj.AllAliases + `"}`
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
