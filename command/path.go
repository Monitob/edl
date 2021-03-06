package command

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	_ "github.com/codegangsta/cli"
	"golang.org/x/text/unicode/norm"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
		pt "path"
)

const PathSeparator = string(filepath.Separator)

type ValidFilenameFunc func(string) bool

func IsFile(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.Mode().IsRegular()
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

// IdenticalFilenames uses unicode normalization
func IdenticalFilenames(filename0, filename1 string) bool {
	return bytes.Equal(norm.NFC.Bytes([]byte(filename0)), norm.NFC.Bytes([]byte(filename1)))
}

func GetFileBasename(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func GetDirContent(path string) ([]string, []string) {
	F, _ := ioutil.ReadDir(path)
	subdirs, filenames := make([]string, 0, len(F)), make([]string, 0, len(F))
	for _, f := range F {
		if f.IsDir() {
			subdirs = append(subdirs, f.Name())
		} else if f.Mode().IsRegular() {
			filenames = append(filenames, f.Name())
		}
	}
	return subdirs, filenames
}

func GetDirFilenames(path string, isValid ValidFilenameFunc) []string {
	F, _ := ioutil.ReadDir(path)
	filenames := make([]string, 0, len(F))
	for _, f := range F {
		if f.Mode().IsRegular() && isValid(f.Name()) {
			filenames = append(filenames, f.Name())
		}
	}
	return filenames
}

func GetDirRawFilenames(path string) []string {
	F, _ := ioutil.ReadDir(path)
	filenames := make([]string, 0, len(F))
	for _, f := range F {
		if f.Mode().IsRegular() {
			filenames = append(filenames, f.Name())
		}
	}
	return filenames
}

func IsResolutionDir(str string) bool {
    if res, _ := regexp.MatchString("DS_Store", str); res == true {
        return false
    }
	fileRegexp := regexp.MustCompile("[[:digit:]]{4}x[[:digit:]]{4}")
	return fileRegexp.MatchString(str)
}

func GetDirSubDir(path string) []string {
	F, _ := ioutil.ReadDir(path)
	subdirs, _ := make([]string, 0, len(F)), make([]string, 0, len(F))

	IsResolutionDir := func(str string) bool {
		fileRegexp := regexp.MustCompile("[[:digit:]]{4}x[[:digit:]]{4}")
		return fileRegexp.MatchString(str)
	}
	for _, f := range F {
		if f.IsDir() && IsResolutionDir(f.Name()) {
			subdirs = append(subdirs, f.Name())
		}
	}
	return subdirs
}

func GetDirSubDirRoot(path string) []string {
	F, _ := ioutil.ReadDir(path)
	subdirs, _ := make([]string, 0, len(F)), make([]string, 0, len(F))

	for _, f := range F {
		if f.IsDir()  {
			subdirs = append(subdirs, pt.Clean(path + "/"  + f.Name()))
		}
	}
	return subdirs
}

func GetCurDirFilenames(isValid ValidFilenameFunc) []string {
	path, _ := os.Getwd()
	return GetDirFilenames(path, isValid)
}

func GetOutputFilename(input, output, ext string) string {
	if output == "" {
		return strings.TrimSuffix(input, filepath.Ext(input)) + ext
	} else if IsDir(output) {
		return output + PathSeparator + GetFileBasename(input) + ext
	}
	return output + ext
}

func GetInputOutput(input, output string, isValid ValidFilenameFunc, ext string) ([]string, []string) {
	var F, G []string
	if input == "" {
		F = GetCurDirFilenames(isValid)
	} else if IsDir(input) {
		F = GetDirFilenames(input, isValid)
	} else {
		F = []string{input}
	}
	for _, f := range F {
		G = append(G, GetOutputFilename(f, output, ext))
	}
	return F, G
}

func GetDirsFromFlagSetArgs(flags *flag.FlagSet) ([]string, error) {
	if flags.NArg() == 0 {
		wd, err := os.Getwd()
		return []string{wd}, err
	}
	var err error
	dirs := make([]string, 0, flags.NArg())
	for _, arg := range flags.Args() {
		if IsDir(arg) {
			dirs = append(dirs, filepath.Clean(arg))
		} else {
			err = fmt.Errorf("%s is not a directory", arg)
		}
	}
	return dirs, err
}

func NewInOut(input, output string) *InOut {
	return &InOut{In: input, Out: output}
}

func CreateDir(folder string) string {
	path, err := os.Getwd()
	if err != nil {
		panic("Error to locate pwd")
	}
	dir := path + "/" + folder
	fmt.Println(dir)
	if _, err := os.Stat(dir); err == nil {
		return dir + "/"
	} else {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			panic("MkdirAll error: Please verify your acces right")
		}
	}
	return dir + "/"
}

func (inout *InOut) Open() error {
	var err error
	if inout.In == "" {
		inout.Reader = bufio.NewReader(os.Stdin)
	} else {
		inout.in, err = os.Open(inout.In)
		if err != nil {
			return err
		}
		inout.Reader = bufio.NewReader(inout.in)
	}
	if inout.Out == "" {
		//os.MkdirAll("DefaulFolderConformation", 777)
	} else {
		// if false != IsDir(inout.Out) {
		// 	os.MkdirAll(inout.Out, 777)
		// }
	}
	return nil
}
