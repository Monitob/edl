package command

import (
	"github.com/codegangsta/cli"
  "fmt"
)


func CmdVerif(c *cli.Context){
    edl := c.String("e")
    FileConf := c.String("d")

    fmt.Println(FileConf)

    if CheckFlags(c.String("e"), c.String("d"), "default") == false {
      return
    }
    fmt.Println(edl)
    fmt.Println(FileConf)
}
