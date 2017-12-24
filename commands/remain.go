package commands

import (
	"fmt"
	"net"

	"github.com/urfave/cli"
)

func Remain(c *cli.Context) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.GlobalString("bind"), c.GlobalInt("port")))
	if err != nil {
		if c.Bool("ignore-error") {
			return nil
		}
		return cli.NewExitError(err, 1)
	}
	defer conn.Close()

	reply := make([]byte, 1024)

	_, err = conn.Read(reply)
	if err != nil {
		if c.Bool("ignore-error") {
			return nil
		}
		return cli.NewExitError(err, 1)
	}

	fmt.Printf("Gomodoro remain %s", reply)
	return nil
}
