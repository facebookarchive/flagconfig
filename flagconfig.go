// Package flagconfig provides a flag to specifiy a config file which
// will in turn be used to read unspecified flag values.
package flagconfig

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var configFile = flag.String(
	"c", defaultConfig(), "Config file to read flags from.")

func Usage() {
	Parse()
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(os.Stderr, "  -%s=%s: %s\n", f.Name, f.Value.String(), f.Usage)
	})
}

func defaultConfig() string {
	home := os.Getenv("HOME")
	basename := filepath.Base(os.Args[0])
	path := filepath.Join(home, ".config", basename, "config")
	_, err := os.Open(path)
	if err == nil {
		return path
	}
	path = filepath.Join("/", "etc", "conf.d", basename)
	_, err = os.Open(path)
	if err == nil {
		return path
	}
	return ""
}

func contains(list []*flag.Flag, f *flag.Flag) bool {
	for _, i := range list {
		if i == f {
			return true
		}
	}
	return false
}

func readConfig() map[string]string {
	bytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file %s: %s", *configFile, err)
	}
	lines := strings.Split(string(bytes), "\n")
	result := make(map[string]string, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed[0] == '#' {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			log.Fatalf("Invalid config line: %s", line)
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result
}

func Parse() {
	if *configFile == "" {
		return
	}
	config := readConfig()
	explicit := make([]*flag.Flag, 0)
	all := make([]*flag.Flag, 0)
	flag.Visit(func(f *flag.Flag) {
		explicit = append(explicit, f)
	})
	flag.VisitAll(func(f *flag.Flag) {
		all = append(all, f)
		if !contains(explicit, f) {
			val := config[f.Name]
			if val != "" {
				err := f.Value.Set(val)
				if err != nil {
					log.Fatalf("Failed to set flag %s with value %s", f.Name, val)
				}
			}
		}
	})
Outer:
	for name, val := range config {
		for _, f := range all {
			if f.Name == name {
				continue Outer
			}
		}
		log.Fatalf("Unknown flag %s=%s in config file.", name, val)
	}
}
