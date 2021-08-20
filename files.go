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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func getCljCacheDir(configDir string) (string, error) {
	env, found := os.LookupEnv("CLJ_CACHE")
	if found {
		return path.Join(env), nil
	}
	env, found = os.LookupEnv("XDG_CACHE_HOME")
	if found {
		return path.Join(env, "clojure"), nil
	}
	return path.Join(configDir, ".cpcache"), nil
}

func getConfigDir() (string, error) {
	env, found := os.LookupEnv("CLJ_CONFIG")
	if found {
		return path.Join(env), nil
	}
	env, found = os.LookupEnv("XDG_CONFIG_HOME")
	if found {
		return path.Join(env, "clojure"), nil
	}
	env, found = os.LookupEnv("HOME")
	if found {
		return path.Join(env, ".clojure"), nil
	}
	env, error := os.UserHomeDir()
	if error != nil {
		return "", error
	}
	return path.Join(env, ".clojure"), nil
}

func getCljToolsDir(configDir string) string {
	return path.Join(configDir, "tools")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func copyFile(dest string, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func checkIsNewerFile(file1 string, file2 string) (bool, error) {
	if !fileExists(file1) {
		return false, nil
	}
	if !fileExists(file2) {
		return true, nil
	}
	sfile1, err := os.Stat(file1)
	if err != nil {
		return false, err
	}
	sfile2, err := os.Stat(file2)
	if err != nil {
		return false, err
	}

	return sfile1.ModTime().UnixNano() > sfile2.ModTime().UnixNano(), nil
}

func pickFiles(toolsDir string, tarPath string, files []string) error {
	tarFile, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	gz, err := gzip.NewReader(tarFile)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		for _, f := range files {
			saveFrom := path.Join("clojure-tools", f)
			if hdr.Name == saveFrom {
				saveTo := path.Join(toolsDir, f)
				if strings.HasSuffix(f, ".jar") {
					saveTo = path.Join(toolsDir, libexecDir, f)
				}
				saveFile, err := os.Create(saveTo)
				if err != nil {
					return err
				}

				if _, err := io.Copy(saveFile, tr); err != nil {
					return err
				}
				fmt.Printf("[t4c] - %s: ", hdr.Name)
				fmt.Printf("... copied to %s", path.Join(f))
				fmt.Println()
			}
		}
	}
	return nil
}

func getExecCpFile(cp string, execJarPath string) string {
	cpFile := cp
	if strings.HasPrefix(cp, "@") {
		cpOriginalFile := strings.TrimPrefix(cp, "@")
		cpFile = cpOriginalFile + ".exec"
		newer, _ := checkIsNewerFile(cpFile, cpOriginalFile)
		if !newer {
			copyFile(cpFile, cpOriginalFile)

			f, _ := os.OpenFile(cpFile, os.O_APPEND|os.O_WRONLY, 0644)
			defer f.Close()
			f.Write([]byte(string(os.PathListSeparator) + execJarPath))
		}
		cpFile = "@" + cpFile
	} else {
		cpFile = cp + string(os.PathListSeparator) + execJarPath
	}
	return cpFile
}
