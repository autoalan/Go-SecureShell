package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/c-bata/go-prompt"
)

var hosts [][]uint8

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}

	for _, value := range hosts {
		s = append(s, prompt.Suggest{
			Text: string(value),
		})
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func getHosts(config []byte) ([][]uint8, error) {
	r := regexp.MustCompile(`(?i)Host\W+\b(?P<host>\S+)\b\s`)
	match := r.FindAllSubmatch(config, -1)

	var named [][]uint8
	for _, value := range match {
		named = append(named, value[1])
	}

	if len(named) == 0 {
		return nil, errors.New("Error: Failed to match any hosts in SSH configuration.")
	}

	return named, nil
}

func loadConfig(path string) ([]byte, error) {
	p := path + filepath.FromSlash("/.ssh/config")

	config, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, errors.New("Error: Failed to open SSH configuration.")
	}

	return config, nil
}

func main() {

	var home string

	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	} else {
		home = os.Getenv("HOME")
	}

	config, err := loadConfig(home)
	if err != nil {
		fmt.Println(err)
		return
	}

	hosts, err = getHosts(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	ssh, err := exec.LookPath("ssh")
	if err != nil {
		fmt.Println("Error: Failed to locate ssh binary.")
	}

	host := prompt.Input("#> ", completer)

	for len(host) == 0 {
		fmt.Println("Select a host or type exit to close.")
		host = prompt.Input("#> ", completer)
	}

	if host == "exit" {
		os.Exit(0)
	}

	cmd := &exec.Cmd{
		Path:   ssh,
		Args:   []string{ssh, host},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Error:", err)
		return
	}
}
