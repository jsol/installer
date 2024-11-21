package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
)

var basePath string

func scriptPath(name, script string) string {
	return path.Join(basePath, name, script+".sh")
}

func hasScript(name, script string) bool {
	if _, err := os.Stat(scriptPath(name, script)); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}

func runScript(name, script string, env []string) (string, error) {
	helpers := path.Join(basePath, "helpers.sh")
	env = append(env, "HELPERS="+helpers)
	execPath := scriptPath(name, script)
	toRun := exec.Command("/bin/bash", execPath)
	toRun.Env = env

	res, err := toRun.Output()

	return string(res), err
}

func sudoRunScript(name, script string, env []string) error {
	helpers := path.Join(basePath, "helpers.sh")
	env = append(env, "HELPERS="+helpers)
	execPath := scriptPath(name, script)
	toRun := exec.Command("sudo", "-E", "/bin/bash", execPath)
	toRun.Env = env

	toRun.Stderr = os.Stderr
	toRun.Stdin = os.Stdin
	toRun.Stdout = os.Stdout

	return toRun.Run()
}

func getVersion(name string) string {
	ver, err := runScript(name, "local-version", []string{"PATH=" + os.Getenv("PATH")})

	if err != nil {
		fmt.Printf("Could not read version for %s: %s (%s)\n", name, err.Error(), ver)
		return ""
	}

	return strings.Trim(string(ver), " \n\r")
}

func getRemoteVersion(name string) string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	env := []string{"ARCH=" + arch, "OS=" + osName}
	ver, err := runScript(name, "remote-version", env)

	if err != nil {
		fmt.Printf("Could not read remote version for %s: %s (%s)\n", name, err.Error(), ver)
		return ""
	}

	return strings.Trim(ver, " \n\r")
}

func dependencies(name string) bool {

	if !hasScript(name, "dependencies") {
		return true
	}

	osName := runtime.GOOS
	arch := runtime.GOARCH
	env := []string{"ARCH=" + arch, "OS=" + osName}
	res, err := runScript(name, "dependencies", env)
	if err != nil {
		fmt.Printf("Dependencies require install:\n%s\n", res)
		return false
	}

	return true
}

func download(name, version string) (bool, string) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	env := []string{"VERSION=" + version, "ARCH=" + arch, "OS=" + osName}
	res, err := runScript(name, "download", env)
	if err != nil {
		fmt.Println("Failed to download", err, res)
		fmt.Printf("Command: %s bash %s/download.sh\n", strings.Join(env, " "), name)
		return false, ""
	}

	return true, strings.Trim(string(res), " \n\r")
}

func install(name, version, arg string) bool {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	err := sudoRunScript(name, "install", []string{"VERSION=" + version, "ARCH=" + arch, "OS=" + osName, "ARG=" + arg})
	if err != nil {
		fmt.Println("Failed to install", err)
		return false
	}

	return true
}

func updateOne(name string) {

	localVer := getVersion(name)

	if localVer == "" {
		fmt.Println("No local version installed of", name)
		return
	}

	remoteVer := getRemoteVersion(name)

	if strings.Trim(remoteVer, "v ") == strings.Trim(localVer, "v ") {
		fmt.Printf("%s is up to date (%s)\n", name, remoteVer)
		return
	}

	fmt.Printf("Updating %s, %s => %s\n", name, localVer, remoteVer)
	_, err := os.Stat(scriptPath(name, "download"))
	if err == nil {
		res, arg := download(name, remoteVer)
		if !res {
			return
		}
		install(name, remoteVer, arg)
	} else {
		fmt.Println(err)
	}
}

func firstInstall(name string) {
	if getVersion(name) != "" {
		fmt.Println(name, "is already installed")
		return
	}

	remoteVer := getRemoteVersion(name)
	fmt.Printf("Installing %s (%s)\n", name, remoteVer)

	dependencies(name)

	_, err := os.Stat(scriptPath(name, "download"))
	if err == nil {
		res, arg := download(name, remoteVer)
		if !res {
			return
		}
		install(name, remoteVer, arg)
	} else {
		fmt.Println(err)
	}

}

func updateAll() {

	entries, err := os.ReadDir(basePath)

	if err != nil {
		fmt.Println("Configuration directory does not exist")
		os.Exit(1)
	}

	var wg sync.WaitGroup

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			updateOne(name)
		}(e.Name())
	}

	wg.Wait()
}

func printUsage() {

	fmt.Println("Usage: installer update\n  installer install [name]")
}

func setupTmp() {
	os.RemoveAll("/tmp/installer")
	os.MkdirAll("/tmp/installer", os.ModePerm)
}

func main() {
	flag.Parse()
	args := flag.Args()

  var ok bool
  basePath, ok = os.LookupEnv("INSTALLER_BASEDIR")
  if !ok {
    fmt.Println("Ensure that the env INSTALLER_BASEDIR points at the root of the install scripts")
  }

	if len(args) == 0 {
		printUsage()
		return
	}

	if args[0] == "update" {
		setupTmp()
		updateAll()
	} else if args[0] == "install" {
		if len(args) != 2 {
			printUsage()
			return
		}
		setupTmp()
		firstInstall(args[1])
	} else {
		printUsage()
	}

}
