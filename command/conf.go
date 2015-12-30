package command

import (
	"bufio"
	"container/list"
	"fmt"
	"github.com/codegangsta/cli"
	_ "io/ioutil"
	"math"
	"os"
	p "path"
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

type Entry struct {
	Event, Reel, TrackType, EditType, Transition string
	SourceIn, SourceOut                          string
	RecordIn, RecordOut                          string
	Notes                                        []string
	TimeIn, TimeOut                              [4]string
	FramesIn, FramesOut                          int
	Elapsed, Seconds, Frames                     int
}

type FilesInfo struct {
	Name      string
	InitFrame int
	EndFrame  int
}

var RegExpEntry = regexp.MustCompile(`^\s*([0-9]+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S*)\s*` +
	`([0-9]{2}:[0-9]{2}:[0-9]{2}:[0-9]{2})\s+` + `([0-9]{2}:[0-9]{2}:[0-9]{2}:[0-9]{2})\s+` +
	`([0-9]{2}:[0-9]{2}:[0-9]{2}:[0-9]{2})\s+` + `([0-9]{2}:[0-9]{2}:[0-9]{2}:[0-9]{2})\s*\*?(.*)`)

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
		} else {
			if entry != nil && isNote(line) {
				entry.Notes = append(entry.Notes, strings.TrimSpace(line[1:]))
			}
		}
	}
	return entries
}

func GetHMS(timecode string) (int, int, int, int) {
	// Return true if 'value' char.
	f := func(c rune) bool {
		return c == ':'
	}
	// Separate into fields with func.
	fields := strings.FieldsFunc(timecode, f)
	// Separate into H M S F  with Fields.
	H, _ := strconv.Atoi(fields[0])
	M, _ := strconv.Atoi(fields[1])
	S, _ := strconv.Atoi(fields[2])
	F, _ := strconv.Atoi(fields[3])
	return H, M, S, F
}

func SplitRawFile(raw string) (NameDir, frames string) {
	f := func(c rune) bool {
		return c == '.'
	}
	fields := strings.FieldsFunc(raw, f)
	if len(fields) > 0 && IsResolutionDir(fields[0]) == false {
		return fields[0], fields[1]
	}
	return "", ""
}

func CheckFlags(edl, dir, root string) bool {
	if len(edl) == 0 {
		fmt.Println("Please specify one edl file")
		return false
	} else if len(dir) == 0 {
		fmt.Println("Please specify one Directory name")
		return false
	} else if len(root) == 0 {
		fmt.Println("Please specify the location to search")
		return false
	}
	return true
}

func CmdConf(c *cli.Context) {

	in := c.String("e")
	out := c.String("p")
	RootFiles := c.String("d")

	if CheckFlags(c.String("e"), c.String("p"), c.String("d")) == false {
		return
	}
	if len(c.Args()) == 1 {
		fmt.Printf("Incorrect usage\n")
		return
	}

	InOut := NewInOut(in, out)

	isEDLFile := func(path string) bool {
		ext := strings.ToLower(filepath.Ext(in))
		return ext == ".edl" || ext == ".txt"
	}

	F, G := GetInputOutput(in, out, isEDLFile, ".conf")
	fmt.Printf("test: %v\n", F)
	fmt.Printf("test: %v\n\n", G)

	_ = InOut.Open()
	defer InOut.in.Close()
	Entry := Parse(InOut, 24)

	IsSrcFolder := func(str string) bool {
		fileRegexp := regexp.MustCompile("^[[:upper:]]{3}[0-9]{3}^C[0-9]{3}|_*[0-9]" + "[^100]")
		return fileRegexp.MatchString(str)
	}

	tc_to_frame := func(timecode string, frame_rate int) int {
		hh, mm, ss, ff := GetHMS(timecode)
		return ff + (ss+mm*60+hh*3600)*frame_rate
	}

	FileList := list.New()
	for _, v := range Entry {
		if IsSrcFolder(v.Reel) == true {
			FileList.PushBack(&FilesInfo{v.Reel, tc_to_frame(v.SourceIn, 24), tc_to_frame(v.SourceOut, 24)})
		}
	}

	// for e := FileList.Front(); e != nil; e = e.Next() {
	// 	fmt.Println(e.Value.(*FilesInfo).Name)
	// 	fmt.Println(e.Value.(*FilesInfo).InitFrame)
	// 	fmt.Println(e.Value.(*FilesInfo).EndFrame)
	// }

	OrgPath := RootFiles
	Recurse := true
	walkFn := func(path string, info os.FileInfo, err error) error {
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}

		IsResolutionDir := func(str string) bool {
			fileRegexp := regexp.MustCompile("[[:digit:]]{4}x[[:digit:]]{4}")
			return fileRegexp.MatchString(str)
		}

		IsInList := func(L *list.List, name, frames string) bool {
			for e := L.Front(); e != nil; e = e.Next() {
				fr, _ := strconv.Atoi(frames)
				if e.Value.(*FilesInfo).Name == name && fr >= e.Value.(*FilesInfo).InitFrame && fr <= e.Value.(*FilesInfo).EndFrame {
					return true
				}
				// fmt.Println(e.Value.(*FilesInfo).Name)
				// fmt.Println(e.Value.(*FilesInfo).InitFrame)
				// fmt.Println(e.Value.(*FilesInfo).EndFrame)
			}
			return false
		}

		if stat.IsDir() && path != OrgPath && !Recurse {
			fmt.Println("skipping dir:", path)
			return filepath.SkipDir
		}

		if err != nil {
			return err
		}
		if IsResolutionDir(path) == true {
			// fmt.Println(path)
			// fmt.Println(p.Base(path))
			name, frames := SplitRawFile(p.Base(path))
			fmt.Println(IsInList(FileList, name, frames), p.Base(path))
		}
		return nil
	}

	err := filepath.Walk(OrgPath, walkFn)
	if err != nil {
		return
	}

}
