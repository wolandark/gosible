package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"gopkg.in/ini.v1"
)

type Server struct {
	IP      string `ini:"ip"`
	User    string `ini:"user"`
	Port    string `ini:"port"`
	KeyPath string `ini:"keyPath"`
}

var s Server
var config string

func readINI(path string) ([]Server, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}
	var servers []Server
	for _, sec := range cfg.Sections() {
		if sec.Name() == ini.DefaultSection || sec.Key("ip").String() == "" {
			continue
		}
		var s Server
		// MapTo does not inherit DEFAULT, so map it first and let the section override.
		if err := cfg.Section(ini.DefaultSection).MapTo(&s); err != nil {
			return nil, err
		}
		if err := sec.MapTo(&s); err != nil {
			return nil, fmt.Errorf("%s: %w", sec.Name(), err)
		}
		if s.KeyPath == "" || s.User == "" || s.Port == "" {
			return nil, fmt.Errorf("%s: missing user, port or keyPath", sec.Name())
		}
		servers = append(servers, s)
	}
	return servers, nil
}

func buildCmd(s Server) string {
	sshCmd := fmt.Sprintf("ssh -i %s -p %s %s@%s", s.KeyPath, s.Port, s.User, s.IP)
	return sshCmd
}

func executeInRemote(s Server, userCmd string) error {
	out, err := exec.Command("sh", "-c", buildCmd(s)+" "+shellQuote(userCmd)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w: %s", s.IP, err, out)
	}
	fmt.Printf("%s", out)
	return nil
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

func main() {
	flag.StringVar(&config, "c", "config.ini", "path to config file")
	flag.Parse()

	uc := strings.Join(flag.Args(), " ")
	if config == "" || uc == "" {
		log.Fatal("usage: gome -c <config.ini> <remote command>")
	}

	servers, err := readINI(config)
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}

	for i, s := range servers {
		fmt.Printf("\n=== Processing IP %d/%d: %s ===\n", i+1, len(servers), s.IP)
		if err := executeInRemote(s, uc); err != nil {
			log.Printf("FAILED %s: %v", s.IP, err)
			continue
		}
		fmt.Printf("Successfully processed %s\n", s.IP)
	}
}
