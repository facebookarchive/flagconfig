// Package flagconfig provides a flag to specifiy a config file which
// will in turn be used to read unspecified flag values.
package flagconfig

import (
	"strings"
	"io/ioutil"
	"log"
	"flag"
)

var configFile = flag.String("c", "", "Config file to read flags from.")

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
	explicit := []*flag.Flag{}
	flag.Visit(func(f *flag.Flag) {
		explicit = append(explicit, f)
	})
	flag.VisitAll(func(f *flag.Flag) {
		if !contains(explicit, f) {
			val := config[f.Name]
			if val != "" {
				f.Value.Set(val)
			}
		}
	})
}
