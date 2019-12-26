package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/decred/dcrd/dcrutil"
	"github.com/jessevdk/go-flags"
)

const (
	defaultConfigFilename = "idleminer.conf"
	defaultDcrNodeHost    = "localhost"
	defaultDcrNodePort    = "19109"
)

var (
	defaultDataDir    = dcrutil.AppDataDir("idleminer", false)
	defaultConfigFile = filepath.Join(
		defaultDataDir, defaultConfigFilename,
	)
)

type config struct {
	NodeHost string `short:"h" long:"node-host" description:"network address of dcrd RPC"`
	Port     string `long:"port" description:"network port of dcrd RPC"`
	User     string `short:"u" long:"user" description:"user of dcrd RPC"`
	Password string `short:"p" long:"password" description:"password of dcrd RPC"`
	Time     int    `short:"t" long:"time" description:"Time in minutes after a mined block to generate new one"`
}

func loadConfig() (*config, []string, error) {
	// Default config
	cfg := config{
		NodeHost: defaultDcrNodeHost,
		Port:     defaultDcrNodePort,
		Time:     10,
	}

	preCfg := cfg
	preParser := flags.NewParser(&preCfg, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return nil, nil, err
		}
	}

	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)

	var configFileError error
	parser := flags.NewParser(&cfg, flags.Default)

	err = flags.NewIniParser(parser).ParseFile(defaultConfigFile)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			fmt.Fprintf(os.Stderr, "Error parsing config "+"file: %v\n", err)
			return nil, nil, err
		}
		configFileError = err
	}

	// Parse command line options again to ensure they take precedence
	remainingArgs, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, usageMessage)
		}
		return nil, nil, err
	}

	// Create the home dir for application
	funcName := "loadConfig"
	err = os.MkdirAll(defaultDataDir, 0700)
	if err != nil {
		// Show error if it's a symlink to diretory does not exist
		if e, ok := err.(*os.PathError); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				err = fmt.Errorf(str, e.Path, link)
			}
		}

		str := "%s: Failed to create home directory: %v"
		err := fmt.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return nil, nil, err
	}

	if configFileError != nil {
		fmt.Printf("%v\n", configFileError)
	}

	return &cfg, remainingArgs, nil
}
