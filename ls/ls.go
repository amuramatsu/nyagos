package ls

import "fmt"
import "io"
import "os"
import "path"
import "strings"
import "../box"

var exeSuffixes = map[string]bool{
	".bat": true,
	".cmd": true,
	".com": true,
	".exe": true,
}

const (
	O_STRIP_DIR = 1
	O_LONG      = 2
	O_INDICATOR = 4
)

func lsOneLong(status os.FileInfo, flag int, out io.Writer) {
	indicator := " "
	if status.IsDir() {
		io.WriteString(out, "d")
		indicator = "/"
	} else {
		io.WriteString(out, "-")
	}
	mode := status.Mode()
	perm := mode.Perm()
	name := status.Name()
	if (perm & 4) > 0 {
		io.WriteString(out, "r")
	} else {
		io.WriteString(out, "-")
	}
	if (perm & 2) > 0 {
		io.WriteString(out, "w")
	} else {
		io.WriteString(out, "-")
	}
	if (perm & 1) > 0 {
		io.WriteString(out, "x")
	} else if exeSuffixes[strings.ToLower(path.Ext(name))] {
		io.WriteString(out, "x")
		indicator = "*"
	} else {
		io.WriteString(out, "-")
	}
	if (flag & O_STRIP_DIR) > 0 {
		name = path.Base(name)
	}
	io.WriteString(out, fmt.Sprintf("%7d %s", status.Size(), name))
	if (flag & O_INDICATOR) != 0 {
		io.WriteString(out, indicator)
	}
	io.WriteString(out, "\n")
}

func lsBox(nodes []os.FileInfo, flag int, out io.Writer) {
	nodes_ := make([]string, len(nodes))
	for key, val := range nodes {
		nodes_[key] = val.Name()
	}
	box.Print(nodes_, 80, out)
}

func lsLong(nodes []os.FileInfo, flag int, out io.Writer) {
	for _, finfo := range nodes {
		lsOneLong(finfo, O_STRIP_DIR|flag, out)
	}
}

type myFileInfo struct {
	nodes []os.FileInfo
}

func (this *myFileInfo) Len() int {
	return len(this.nodes)
}
func (this *myFileInfo) Less(i, j int) bool {
	return this.nodes[i].Name() < this.nodes[j].Name()
}
func (this *myFileInfo) Swap(i, j int) {
	tmp := this.nodes[i]
	this.nodes[i] = this.nodes[j]
	this.nodes[j] = tmp
}

func lsFolder(folder string, flag int, out io.Writer) error {
	fd, err := os.Open(folder)
	if err != nil {
		return err
	}
	defer fd.Close()
	var nodesArray myFileInfo
	nodesArray.nodes, err = fd.Readdir(-1)
	if err != nil {
		return err
	}
	if (flag & O_LONG) > 0 {
		lsLong(nodesArray.nodes, flag, out)
	} else {
		lsBox(nodesArray.nodes, flag, out)
	}
	return nil
}

func lsCore(paths []string, flag int, out io.Writer) error {
	if len(paths) <= 0 {
		return lsFolder(".", flag, out)
	}
	dirs := make([]string, 0)
	printCount := 0
	for _, name := range paths {
		status, err := os.Stat(name)
		if err != nil {
			return err
		}
		if status.IsDir() {
			dirs = append(dirs, name)
		} else {
			lsOneLong(status, flag, out)
			printCount += 1
		}
	}
	for _, name := range dirs {
		if len(paths) > 1 {
			if printCount > 0 {
				io.WriteString(out, "\n")
			}
			io.WriteString(out, name)
			io.WriteString(out, ":\n")
		}
		err := lsFolder(name, flag, out)
		if err != nil {
			return err
		}
		printCount++
	}
	return nil
}

var option = map[rune](func(*int) error){
	'l': func(flag *int) error {
		*flag |= O_LONG
		return nil
	},
	'F': func(flag *int) error {
		*flag |= O_INDICATOR
		return nil
	},
}

// 存在しないオプションに関するエラー
type OptionError struct {
	Option rune
}

func (this OptionError) Error() string {
	return fmt.Sprintf("-%c: No such option", this.Option)
}

// ls 機能のエントリ:引数をオプションとパスに分離する
func Main(args []string, out io.Writer) error {
	flag := 0
	paths := make([]string, 0)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			for _, o := range arg[1:] {
				setter, ok := option[o]
				if !ok {
					var err OptionError
					err.Option = o
					return err
				}
				err := setter(&flag)
				if err != nil {
					return err
				}
			}
		} else {
			paths = append(paths, arg)
		}
	}
	return lsCore(paths, flag, out)
}

// vim:set fenc=utf8 ts=4 sw=4 noet:
