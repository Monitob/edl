package command

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type InOut struct {
	bufio.ReadWriter
	In, Out string
	in, out *os.File
}

var RegExpEntry = regexp.MustCompile(`^\s*([0-9]+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S*)\s*` +
	`([0-9]{2}:[0-9]{2}:[0-9]{2}:[0-9]{2})\s+` + `([0-9]{2}:[0-9]{2}:[0-9]{2}:[0-9]{2})\s+` +
	`([0-9]{2}:[0-9]{2}:[0-9]{2}:[0-9]{2})\s+` + `([0-9]{2}:[0-9]{2}:[0-9]{2}:[0-9]{2})\s*\*?(.*)`)

type Entry struct {
	Event, Reel, TrackType, EditType, Transition string
	SourceIn, SourceOut                          string
	RecordIn, RecordOut                          string
	Notes                                        []string
	TimeIn, TimeOut                              [4]string
	FramesIn, FramesOut                          int
	Elapsed, Seconds, Frames                     int
}

func NewEntry(S []string, fps int) *Entry {
	e := &Entry{Notes: make([]string, 0, 10)}
	e.Event, e.Reel, e.TrackType, e.EditType, e.Transition = S[0], S[1], S[2], S[3], S[4]
	e.SourceIn, e.SourceOut, e.RecordIn, e.RecordOut = S[5], S[6], S[7], S[8]

	if S[9] != "" {
		e.Notes = append(e.Notes, strings.TrimSpace(S[9]))
	}
	var time [4]int
	for i, s := range strings.Split(e.RecordIn, ":") {
		time[i], _ = strconv.Atoi(s)
		e.TimeIn[i] = s
	}
	e.FramesIn = time[0]*60*60*fps + time[1]*60*fps + time[2]*fps + time[3]
	for i, s := range strings.Split(e.RecordOut, ":") {
		time[i], _ = strconv.Atoi(s)
		e.TimeOut[i] = s
	}
	e.FramesOut = time[0]*60*60*fps + time[1]*60*fps + time[2]*fps + time[3]
	e.Elapsed = e.FramesOut - e.FramesIn
	e.Seconds = e.Elapsed / fps
	e.Frames = int(math.Mod(float64(e.Elapsed), float64(fps)))
	return e
}

func Parse(r *InOut, fps int) []*Entry {

	// Set the split function for the scanning operation.
	scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		innerline, endline := regexp.MustCompile("\r([^\n])"), regexp.MustCompile("\r$")
		replaced := endline.ReplaceAll(innerline.ReplaceAll(data, []byte("\n$1")), []byte("\n"))

		return bufio.ScanLines(replaced, atEOF)
	}

	// notes start with a `*`
	isNote := func(s string) bool {
		return len(s) > 0 && s[0] == '*'
	}

	var entry *Entry
	var entries []*Entry
	scanner := bufio.NewScanner(r.in)
	scanner.Split(scanLines)
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println(line)
		if S := RegExpEntry.FindStringSubmatch(line); S != nil {
			entry = NewEntry(S[1:], fps)
			entries = append(entries, entry)
			fmt.Printf("Reel %v\n", entry.Reel)
			fmt.Printf("Source in %v\n", entry.SourceIn)
			fmt.Printf("Source out %v\n", entry.SourceOut)
		} else {
			if entry != nil && isNote(line) {
				entry.Notes = append(entry.Notes, strings.TrimSpace(line[1:]))
			}
		}
	}
	return entries
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
	Entry := Parse(InOut, 24)
	ScanFiles := func(str string) bool {
		fileRegexp := regexp.MustCompile("[^A-Z]{3}&[0-9]{3}C[0-9]{3}|_*[0-9]")
		return fileRegexp.MatchString(str)
	}
	for _, v := range Entry {
		fmt.Println(ScanFiles(v.Reel), v.Reel)
	}
}
