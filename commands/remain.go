package commands

import (
	"fmt"
	"net"

	"github.com/urfave/cli"
)

// Remain UNIX Socketを利用して残り時間を標準出力する
func Remain(c *cli.Context) error {
	conn, err := net.Dial("unix", c.GlobalString("socket-path"))
	if err != nil {
		if c.Bool("ignore-error") {
			fmt.Printf("--:--")
			return nil
		}
		return cli.NewExitError(err, 1)
	}
	defer conn.Close()

	reply := make([]byte, 1024)

	_, err = conn.Read(reply)
	if err != nil {
		if c.Bool("ignore-error") {
			fmt.Printf("--:--")
			return nil
		}
		return cli.NewExitError(err, 1)
	}

	fmt.Printf("%s", reply)
	return nil
}
