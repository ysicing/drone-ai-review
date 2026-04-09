package main

import (
	"os"

	"github.com/drone-stack/drone-plugin-template/plugin"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version = "0.0.1"
)

type formatter struct{}

func (*formatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}

func init() {
	logrus.SetFormatter(new(formatter))
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		_ = godotenv.Load(env)
	}

	app := cli.NewApp()
	app.Name = "drone example plugin"
	app.Usage = "drone example plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug",
			EnvVar: "PLUGIN_DEBUG",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := plugin.Plugin{
		Ext: plugin.Ext{
			Debug: c.Bool("debug"),
		},
	}

	if plugin.Ext.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := plugin.Exec(); err != nil {
		logrus.Fatal(err)
	}
	return nil
}
