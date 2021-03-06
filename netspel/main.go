package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/myshkin5/netspel/adapters/sse"
	"github.com/myshkin5/netspel/adapters/udp"
	"github.com/myshkin5/netspel/factory"
	"github.com/myshkin5/netspel/logs"
	"github.com/myshkin5/netspel/schemes/simple"
	"github.com/myshkin5/netspel/schemes/streaming"
	"github.com/op/go-logging"
)

func init() {
	factory.WriterManager.RegisterType("udp", reflect.TypeOf(udp.Writer{}))
	factory.ReaderManager.RegisterType("udp", reflect.TypeOf(udp.Reader{}))

	factory.WriterManager.RegisterType("sse", reflect.TypeOf(sse.Writer{}))
	factory.ReaderManager.RegisterType("sse", reflect.TypeOf(sse.Reader{}))

	factory.SchemeManager.RegisterType("simple", reflect.TypeOf(simple.Scheme{}))
	factory.SchemeManager.RegisterType("streaming", reflect.TypeOf(streaming.Scheme{}))
}

func main() {
	app := cli.NewApp()
	app.Name = "netspel"
	app.Usage = "test network throughput with varying protocols"
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "configuration file",
		},
		cli.StringFlag{
			Name:  "scheme, s",
			Usage: "scheme type overriding the config file",
		},
		cli.StringFlag{
			Name:  "writer, w",
			Usage: "writer type overriding the config file",
		},
		cli.StringFlag{
			Name:  "reader, r",
			Usage: "reader type overriding the config file",
		},
		cli.StringSliceFlag{
			Name:  "config-string",
			Usage: "additional configuration <key>=<value> strings overriding the config file",
		},
		cli.StringSliceFlag{
			Name:  "config-int",
			Usage: "additional configuration <key>=<value> integers overriding the config file",
		},
		cli.StringFlag{
			Name:   "log-level, l",
			Usage:  "logging level",
			EnvVar: "INFO,DEBUG",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:    "write",
			Aliases: []string{"w"},
			Usage:   "write messages",
			Action: func(context *cli.Context) {
				write(context)
			},
		},
		cli.Command{
			Name:    "read",
			Aliases: []string{"r"},
			Usage:   "read messages",
			Action: func(context *cli.Context) {
				read(context)
			},
		},
	}

	app.RunAndExitOnError()
}

func write(context *cli.Context) {
	initLogs(context)

	config := config(context)
	scheme := scheme(config, context)

	writer, err := factory.CreateWriter(config.WriterType)
	if err != nil {
		panic(err)
	}

	err = writer.Init(config.Additional)
	if err != nil {
		cli.ShowAppHelp(context)
		panic(err)
	}

	scheme.RunWriter(writer)
}

func read(context *cli.Context) {
	initLogs(context)

	config := config(context)
	scheme := scheme(config, context)

	reader, err := factory.CreateReader(config.ReaderType)
	if err != nil {
		cli.ShowAppHelp(context)
		panic(err)
	}

	err = reader.Init(config.Additional)
	if err != nil {
		panic(err)
	}

	scheme.RunReader(reader)
}

func initLogs(context *cli.Context) {
	level, err := logging.LogLevel(context.GlobalString("log-level"))
	if err != nil {
		level = logging.INFO
	}

	logs.LogLevel.SetLevel(level, "netspel")
}

func config(context *cli.Context) factory.Config {
	configPath := context.GlobalString("config")
	var config factory.Config
	var err error
	if configPath == "" {
		config, err = factory.Parse([]byte("{}"))
	} else {
		config, err = factory.LoadFromFile(configPath)
	}
	if err != nil {
		cli.ShowAppHelp(context)
		panic(err)
	}

	schemeType := context.GlobalString("scheme")
	if schemeType != "" {
		config.SchemeType = schemeType
	}
	writerType := context.GlobalString("writer")
	if writerType != "" {
		config.WriterType = writerType
	}
	readerType := context.GlobalString("reader")
	if readerType != "" {
		config.ReaderType = readerType
	}

	for _, assignment := range context.GlobalStringSlice("config-string") {
		keyValue, err := parseAssignment(assignment)
		if err != nil {
			panic(err)
		}

		config.Additional.SetString(keyValue[0], keyValue[1])
	}
	for _, assignment := range context.GlobalStringSlice("config-int") {
		keyValue, err := parseAssignment(assignment)
		if err != nil {
			panic(err)
		}

		value, err := strconv.Atoi(keyValue[1])
		if err != nil {
			panic(err)
		}

		config.Additional.SetInt(keyValue[0], value)
	}

	return config
}

func parseAssignment(assignment string) ([]string, error) {
	keyValue := strings.Split(assignment, "=")
	if len(keyValue) != 2 {
		return []string{}, fmt.Errorf("Values must be of the form <key>=<value>, %s", assignment)
	}

	return keyValue, nil
}

func scheme(config factory.Config, context *cli.Context) factory.Scheme {
	scheme, err := factory.CreateScheme(config.SchemeType)
	if err != nil {
		cli.ShowAppHelp(context)
		panic(err)
	}

	err = scheme.Init(config.Additional)
	if err != nil {
		panic(err)
	}

	return scheme
}
