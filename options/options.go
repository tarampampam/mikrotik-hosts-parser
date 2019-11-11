package options

import (
	"errors"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"strings"
)

type Options struct {
	Address          string `short:"l" long:"listen" env:"LISTEN_ADDR" default:"0.0.0.0" description:"Address (IP) to listen on"`
	Port             int    `short:"p" long:"port" env:"LISTEN_PORT" default:"8080" description:"TCP port number"`
	ResourcesDir     string `long:"resources-dir" env:"RESOURCES_DIR" description:"Resources directory path"`
	IndexFileName    string `long:"index-name" env:"INDEX_NAME" default:"index.html" description:"Index file name"`
	Error404FileName string `long:"error404-name" env:"ERROR404_NAME" default:"404.html" description:"Error 404 file name"`
	ShowVersion      bool   `short:"V" long:"version" description:"Show version and exit"`
	stdLog           *log.Logger
	errLog           *log.Logger
	onExit           ExitFunc
	parseFlags       flags.Options
	Version 		 string
}

type ExitFunc func(code int)

// Create new options instance.
func NewOptions(stdOut, stdErr *log.Logger, onExit ExitFunc) *Options {
	if onExit == nil {
		onExit = func(code int) {
			os.Exit(code)
		}
	}
	return &Options{
		stdLog:     stdOut,
		errLog:     stdErr,
		onExit:     onExit,
		parseFlags: flags.Default,
	}
}

// Parse options using fresh parser instance.
func (o *Options) Parse() *flags.Parser {
	var parser = flags.NewParser(o, o.parseFlags)
	var _, err = parser.Parse()

	// parse passed options
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			o.onExit(0)
		} else {
			parser.WriteHelp(o.stdLog.Writer())
			o.onExit(1)
		}
	}

	// show application version and exit, if flag `-V` passed
	if o.ShowVersion == true {
		o.stdLog.Println("Version: " + o.Version)
		o.onExit(0)
	}

	// make options check
	if _, err := o.Check(); err != nil {
		o.errLog.Println(err.Error())
		o.onExit(1)
	}

	return parser
}

// Make options check.
func (o *Options) Check() (bool, error) {
	// check address
	if len(strings.TrimSpace(o.Address)) < 7 {
		return false, errors.New("wrong address to listen on")
	}

	// check port
	if o.Port <= 0 || o.Port > 65535 {
		return false, errors.New("wrong port number")
	}

	// check resources directory (if defined)
	if len(o.ResourcesDir) > 0 {
		if info, err := os.Stat(o.ResourcesDir); err != nil || !info.Mode().IsDir() {
			return false, errors.New("wrong resources directory")
		}
	}

	return true, nil
}
