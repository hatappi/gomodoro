package main

import (
	"fmt"
	"os"

	"github.com/hatappi/gomodoro/commands"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

var (
	homeDir string
	err     error
)

func init() {
	homeDir, err = homedir.Dir()
	if err != nil {
		panic(err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Gomodoro"
	app.Usage = "Pomodoro Technique By Go"
	app.Version = "0.2.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "conf-path, c",
			Value: fmt.Sprintf("%s/.gomodoro/config.toml", homeDir),
			Usage: "gomodoro config path",
		},
		cli.StringFlag{
			Name:  "app-dir, a",
			Value: fmt.Sprintf("%s/.gomodoro", homeDir),
			Usage: "application directory",
		},
		cli.StringFlag{
			Name:  "socket-path, s",
			Value: "/tmp/gomodoro.sock",
			Usage: "gomodoro socket path",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "pomodoro start",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "long-break-sec, l",
					Value: 15 * 60,
					Usage: "long break (s)",
				},
				cli.IntFlag{
					Name:  "short-break-sec, s",
					Value: 5 * 60,
					Usage: "short break (s)",
				},
				cli.IntFlag{
					Name:  "work-sec, w",
					Value: 25 * 60,
					Usage: "work (s)",
				},
			},
			Action: commands.Start,
		},
		{
			Name:  "remain",
			Usage: "Get Remain",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "ignore-error, i",
				},
			},
			Action: commands.Remain,
		},
	}

	app.Run(os.Args)
}
