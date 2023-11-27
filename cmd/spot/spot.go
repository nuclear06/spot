package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/ssh-honeypot/pkg/config"
	. "github.com/ssh-honeypot/pkg/logging"
	"github.com/ssh-honeypot/pkg/server"
	"github.com/urfave/cli/v2"
	"os"
)

func cliF(c *cli.Context) error {
	// if not pass any args, print hint
	if c.Args().Len() == 0 {
		LogInfo(SysInit, "Not pass any args, use default config, you can use -h to get manual")
	}
	confPath := c.Path("config")
	err := checkConfig(&confPath)
	if err != nil {
		return err
	}

	conf, err := config.ParseConfig(confPath)
	if err != nil {
		return err
	}

	// cli args -p will overwrite port
	if c.Int("port") != 0 {
		conf.Addr = fmt.Sprintf("127.0.0.1:%d", c.Int("port"))
	}
	// cli args -v will overwrite debug mode
	conf.Log.IsDebug = c.Bool("debug") || conf.Log.IsDebug

	LogDebug(SysInit, fmt.Sprintf("Config: %v", conf))
	InitLog(*conf)
	err = server.InitSSH(*conf)
	return err
}

func checkConfig(p *string) error {
	if *p == "" {
		*p = "config.yml"
		LogDebug(SysInit, "config file path is not specific, use default value")
	}
	LogDebug(SysInit, fmt.Sprintf("set config file path to: ./%s", *p))

	t, err := os.Stat(*p)
	if os.IsNotExist(err) {
		LogError(SysInit, fmt.Sprintf("%s is not exist", *p))
		return err
	} else if t.IsDir() {
		LogError(SysInit, fmt.Sprintf("%s it is a directory", *p))
		return fmt.Errorf("%s it is a directory", *p)
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:                   "spot",
		EnableBashCompletion:   true,
		HideHelpCommand:        true,
		UseShortOptionHandling: true,
		Suggest:                true,
		Usage:                  "a simple ssh honeypot",
		Action:                 cliF,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "config",
				Aliases: []string{"d"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"v"},
				Usage:   "debug mode",
				Value:   false,
				Action: func(c *cli.Context, b bool) error {
					if b {
						SetLevel(logrus.DebugLevel)
					}
					return nil
				},
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "listen port",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "conf",
				Usage:  "config tools",
				Action: nil,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "init",
						Aliases: []string{"i"},
						Value:   false,
						Usage:   "generate config template",
						Action: func(c *cli.Context, b bool) error {
							err := config.GenConf(config.DefaultConfig(), "config.yml")
							if err != nil {
								return err
							}
							LogInfo(SysInit, "config file has generated: ./config.yml")
							os.Exit(0)
							return nil
						},
					},
					&cli.PathFlag{
						Name:    "check",
						Aliases: []string{"c"},
						Usage:   "load config file, print runtime config",
						Action: func(c *cli.Context, p string) error {
							parseConfig, err := config.ParseConfig(p)
							if err != nil {
								return err
							}
							marshal, err := json.MarshalIndent(parseConfig, "", "   ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshal))
							os.Exit(0)
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		LogError(SysInit, err.Error())
	}
}
