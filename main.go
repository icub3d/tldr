package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Path is a colon separated list of directories to look for tldr files.
var paths string

// Pull signals the program to update all the tldrs to their latest.
var pull bool

// PullDir is the directory when we put tldrs when pulling them.
var pullDir string

func init() {
	flag.BoolVar(&pull, "pull", false, "pull tldrs from the repo into $HOME/.local/share/tldr.")
	flag.StringVar(&pullDir, "pull-dir", "$HOME/.local/share/tldr",
		"the paths to put tldrs in when pulling.")
	flag.StringVar(&paths, "paths", "$HOME/.tldr:(pull-dir)",
		"the paths to look for tldrs.")
}

func main() {
	// Parse the flags, including replacing (pull-dir) in the path value.
	flag.Parse()
	paths = strings.Replace(paths, "(pull-dir)", pullDir, -1)

	// If we got a pull request, expand it and do the pull.
	pullDir = os.ExpandEnv(pullDir)
	paths = os.ExpandEnv(strings.Replace(paths, "~", "$HOME", -1))
	if pull {
		doPull()
	}

	// We expand the paths and look for the value.
	paths = os.ExpandEnv(paths)
	tldr(flag.Args())
}

// doPull performs the actual pull operation. It grabs data from github.
func doPull() {
	// Make the pull directory so it's there when we try to place files.
	err := os.MkdirAll(pullDir, 0750)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed create tldr directory: %v\n", err)
		os.Exit(1)

	}

	// Get the list of tldrs from github.
	files := []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"download_url"`
	}{}
	URL := "https://api.github.com/repos/chrisallenlane/cheat/contents/cheat/cheatsheets"
	resp, err := http.Get(URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed getting tldr listing: %v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&files); err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing tldr listing: %v\n", err)
		os.Exit(1)
	}

	// Save each of the tldrs by name.
	for _, file := range files {
		getFile(file.Name, file.DownloadURL)
	}
}

// getFile writes the contents of the given url to a file with the
// given name under the pullDir directory.
func getFile(name, URL string) {
	// Get the contents.
	resp, err := http.Get(URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed getting tldr %v: %v\n", name, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Get the file path.
	p := filepath.Join(pullDir, name)
	file, err := os.Create(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed opening file for tldr %v: %v\n", p, err)
		os.Exit(1)
	}
	defer file.Close()

	// Copy.
	if _, err = io.Copy(file, resp.Body); err != nil {
		fmt.Fprintf(os.Stderr, "failed saving tldr %v: %v\n", name, err)
		os.Exit(1)
	}
}

// tldr looks for a tldr for each of the given names.
func tldr(names []string) {
	found := false
	for _, name := range names {
		// Look for the name in each of the paths.
		for _, path := range strings.Split(paths, ":") {
			fullpath := filepath.Join(path, name)
			if _, err := os.Stat(fullpath); err == nil {
				file, err := os.Open(fullpath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to open %v: %v", fullpath, err)
					continue
				}
				if _, err = io.Copy(os.Stdout, file); err != nil {
					fmt.Fprintf(os.Stderr, "failed writing tldr %v to stdout %v: %v\n",
						fullpath, name, err)
				}
				found = true
			}
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "didn't find any tldrs. have you tried 'tldr -pull'?\n")
	}
}
