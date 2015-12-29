package command

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type InOut struct {
	bufio.ReadWriter
	In, Out string
	in, out *os.File
}

func CmdConf(c *cli.Context) {

	in := c.Args().First()
	out := c.Args().Tail()
	InOut := NewInOut(in, out[0])
	isEDLFile := func(path string) bool {
		ext := strings.ToLower(filepath.Ext(c.Args().First()))
		return ext == ".edl" || ext == ".txt"
	}

	F, G := GetInputOutput(in, out[0], isEDLFile, ".conf")
	fmt.Printf("test: %v\n", F)
	fmt.Printf("test: %v\n\n", G)

	_ = InOut.Open()
	scanner := bufio.NewScanner(InOut.in)

	validValue, _ := regexp.Compile(`[0-9]+\s+(A[0-9]{3}_*C[0-9]{3}_*[A-Z0-9]{6}_*[0-9]*)\s+V\s+.*`)
	count := 0
	for scanner.Scan() {
		fmt.Printf("no.- %d  ", count)
		res := validValue.FindAllString(scanner.Text(), -1)
		fmt.Printf(" %v \n ", res)

		count++
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
