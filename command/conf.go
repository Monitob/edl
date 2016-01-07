package command

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	_ "io/ioutil"
	"math"
	"os"
	"os/exec"
	p "path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

type result struct {
	path string
	err  error
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
		cpCmd := exec.Command("cp", src, dst)
		err = cpCmd.Start()
		if err != nil {
			fmt.Printf("Command finished with error: %v\n", err)
			return err
		}
		err = cpCmd.Wait()
		if err != nil {
			fmt.Printf("Command finished with error: %v\n", err)
			return err
		}
	//sfi, err := os.Stat(src)
	// if err != nil {
	// 	return
	// }
	// if !sfi.Mode().IsRegular() {
	// 	// cannot copy non-regular files (e.g., directories,
	// 	// symlinks, devices, etc.)
	// 	return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	// }
	// dfi, err := os.Stat(dst)
	// if err != nil {
	// 	if !os.IsNotExist(err) {
	// 		return
	// 	}
	// } else {
	// 	if !(dfi.Mode().IsRegular()) {
	// 		return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
	// 	}
	// 	if os.SameFile(sfi, dfi) {
	// 		return
	// 	}
	// }

	// err = copyFileContents(src, dst)
	return nil
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
	if len(fields) > 0 {
        if IsResolutionDir(fields[0]) == false {
		    return fields[0], fields[1]
        }
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

func IsSrcFolder(str string) bool {
	fileRegexp := regexp.MustCompile("^[[:upper:]]{3}[0-9]{3}^C[0-9]{3}|_*[0-9]" + "[^100]")
	return fileRegexp.MatchString(str)
}

func tc_to_frame(timecode string, frame_rate int) int {
	hh, mm, ss, ff := GetHMS(timecode)
	return ff + (ss+mm*60+hh*3600)*frame_rate
}

func IsInList(L *list.List, name, frames string) bool {
	for e := L.Front(); e != nil; e = e.Next() {
		fr, _ := strconv.Atoi(frames)
		if e.Value.(*FilesInfo).Name == name && fr >= e.Value.(*FilesInfo).InitFrame && fr <= e.Value.(*FilesInfo).EndFrame {
			return true
		}
	}
	return false
}

// sumFiles starts goroutines to walk the directory tree at root and digest each
// regular file.  These goroutines send the results of the digests on the result
// channel and send the result of the walk on the error channel.  If done is
// closed, sumFiles abandons its work.

func sumFiles(done <-chan struct{}, root string, FileList *list.List) (<-chan result, <-chan error) {
	// For each regular file, start a goroutine that sums the file and sends
	// the result on c.  Send the result of the walk on errc.
	c := make(chan result)
	errc := make(chan error, 1)
	Recurse := true
	go func() { // HL
		var Wg sync.WaitGroup
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
/*			if !info.Mode().IsRegular() {
				return nil
			}*/
			stat, err := os.Stat(path)
            fmt.Println(path)
		 	if err != nil {
                fmt.Println(err)
		 			return err
			}
			if stat.IsDir() && path != root && !Recurse {
			 			fmt.Println("skipping dir:", path)
			 			return filepath.SkipDir
		 		}
			Wg.Add(1)
			go func() {
				if IsResolutionDir(path) == true {
					name, frames := SplitRawFile(p.Base(path))
					if IsInList(FileList, name, frames) == true {
						select {
						case c <- result{path, nil}: // HL
						case <-done: // HL
						}
					}
				}
				Wg.Done()
			}()
			// Abort the walk if done is closed.
			select {
			case <-done: // HL
				return errors.New("walk canceled")
			default:
				return nil
			}
		})
		// Walk has returned, so all calls to Wg.Add are done.  Start a
		// goroutine to close c once all the sends are done.
		go func() { // HL
			Wg.Wait()
			close(c) // HL
		}()
		// No select needed here, since errc is buffered.
		errc <- err // HL
	}()
	return c, errc
}


func CmdConf(c *cli.Context) {

	in := c.String("e")
	out := c.String("p")
	RootLocation := c.String("d")
	desteny := CreateDir(out)
	//Recurse := true
	done := make(chan struct{}) // HLdone
	defer close(done)           // HLdone

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
	fmt.Println(desteny)

	_ = InOut.Open()
	defer InOut.in.Close()
	Entry := Parse(InOut, 24)

	FileList := list.New()
	for _, v := range Entry {
		if IsSrcFolder(v.Reel) == true {
			FileList.PushBack(&FilesInfo{v.Reel, tc_to_frame(v.SourceIn, 24), tc_to_frame(v.SourceOut, 24)})
		}
	}

    subdir := GetDirSubDirRoot(RootLocation)
		    fmt.Println(RootLocation)

    for index, s :=  range subdir {
        cr, _ := sumFiles(done, s, FileList)
	    for r := range cr { // HLrange
		    fmt.Println(r.path, index)
	    	cpCmd := exec.Command("cp", "-rf", r.path, desteny)
						err := cpCmd.Start()
		 				if err != nil {
							fmt.Printf("Commad finished with error: %v\n", err)
		 				}
		 				err = cpCmd.Wait()
		 				if err != nil {
		 					fmt.Printf("Command finished with error: %v\n", err)
		 				}
		    if r.err != nil {
		    	return
		    }
	    }
    }

}
