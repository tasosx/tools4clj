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
	"os/signal"
	"runtime"
	"syscall"
)

func makeClassPathCmd(conf *t4cConfig, toolsClassPath string) exec.Cmd {
	cmdArgs := append([]string{}, "-classpath", toolsClassPath,
		"clojure.main",
		"-m", "clojure.tools.deps.alpha.script.make-classpath2",
		"--config-user", conf.configUser,
		"--config-project", conf.configProject,
		"--libs-file", conf.libsFile,
		"--basis-file", conf.basisFile,
		"--cp-file", conf.cpFile,
		"--jvm-file", conf.jvmFile,
		"--main-file", conf.mainFile)
	cmdArgs = append(cmdArgs, conf.toolsArgs...)

	var cmd = exec.Command(javaPath, cmdArgs...)

	cmd.Args = removeEmpty(cmd.Args)

	return *cmd
}

func generatePomCmd(conf *t4cConfig, toolsClassPath string) exec.Cmd {
	cmdArgs := append([]string{}, "-classpath", toolsClassPath,
		"clojure.main",
		"-m", "clojure.tools.deps.alpha.script.generate-manifest2",
		"--config-user", conf.configUser,
		"--config-project", conf.configProject,
		"--gen=pom")
	cmdArgs = append(cmdArgs, conf.toolsArgs...)

	var cmd = exec.Command(javaPath, cmdArgs...)

	cmd.Args = removeEmpty(cmd.Args)

	return *cmd
}

func clojureExecuteCmd(jvmCacheOpts []string, jvmOpts []string, basisFile string,
	execJarPath string, cp string, args []string) exec.Cmd {

	cmdArgs := append([]string{}, jvmCacheOpts...)
	cmdArgs = append(cmdArgs, jvmOpts...)
	cmdArgs = append(cmdArgs, "-Dclojure.basis="+basisFile,
		"-classpath", getExecCpFile(cp, execJarPath))
	cmdArgs = append(cmdArgs, "clojure.main", "-m", "clojure.run.exec")
	cmdArgs = append(cmdArgs, args...)

	var cmd = exec.Command(javaPath, cmdArgs...)

	cmd.Args = removeEmpty(cmd.Args)

	return *cmd
}

func clojureCmd(jvmCacheOpts []string, jvmOpts []string, libsFile string, basisFile string,
	cp string, mainCacheOpts []string, clojureArgs []string, rlwrap bool) exec.Cmd {

	cmdArgs := append([]string{}, jvmCacheOpts...)
	cmdArgs = append(cmdArgs, jvmOpts...)
	cmdArgs = append(cmdArgs, "-Dclojure.libfile="+libsFile, "-Dclojure.basisfile="+basisFile, "-classpath", cp, "clojure.main")
	cmdArgs = append(cmdArgs, mainCacheOpts...)
	cmdArgs = append(cmdArgs, clojureArgs...)

	var cmd = exec.Command(javaPath, cmdArgs...)

	cmd.Args = removeEmpty(cmd.Args)

	if rlwrap && runtime.GOOS != "windows" {
		rlwrapPath := rlwrapPath()
		if rlwrapPath != "" {
			cmd.Args = append(rlwrapArgs(), cmd.Args...)
			cmd.Path = rlwrapPath
		}
	}

	return *cmd
}

func start(cmd exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

// to safely exit this process,
// catch brake signals and,
// allow started process to exit
func safeStart(cmd exec.Cmd) error {
	done := make(chan error)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		var err = cmd.Start()

		if err != nil {
			done <- err
			return
		}
		err = cmd.Wait()
		if err != nil {
			done <- err
			return
		}
		done <- nil
	}()

	return <-done
}
