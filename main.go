package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

type FileStatus struct {
	LastModTime int64
	ModTime     int64
}
type Watcher struct {
	Directory string   `json:"directory"`
	Excludes  []string `json:"excludes"`
	JobsC     chan<- Job
	Targets   map[string]map[string]FileStatus
}

type Config struct {
	Watchers []*Watcher `json:"watchers"`
}

type Job struct {
	Kind string
	Path string
}

func main() {
	configFile := flag.String("c", "config.json", "Config File")
	interval := flag.Int("i", 3000, "An interval(ms) of a watchers' file check loop")
	flag.Parse()

	bytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		fmt.Errorf("An error occurred: %s \n\n", err.Error())
		os.Exit(-1)
	}

	config := &Config{}
	if err := json.Unmarshal(bytes, config); err != nil {
		fmt.Errorf("An error occurred: %s \n\n", err.Error())
		os.Exit(-1)
	}

	jobsC := make(chan Job, len(config.Watchers))

	wd, err := os.Getwd()
	if err != nil {
		fmt.Errorf("An error occurred: %s \n\n", err.Error())
		os.Exit(-1)
	}

	for _, watcher := range config.Watchers {
		go watcher.Launch(wd, *interval, jobsC)
	}
	handleJobs(jobsC)
}

func getMtime(path string) (mtime time.Time, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	mtime = fi.ModTime()
	return
}
func setMtime(path string, mtime time.Time) (err error) {
	atime := time.Now()
	err = os.Chtimes(path, atime, mtime)
	return
}

func handleJobs(jobsC <-chan Job) {
	for job := range jobsC {
		if job.Kind != "deleted" {
			mTime, err := getMtime(job.Path)
			if err != nil {
				fmt.Errorf("An error occurred touch: %s \n\n", err.Error())
				os.Exit(-1)
			}
			setMtime(job.Path, mTime.Add(2))
		}
	}
}

func (w *Watcher) Launch(watchDir string, interval int, jobsC chan<- Job) {
	w.JobsC = jobsC
	w.Targets = make(map[string]map[string]FileStatus)
	targetDir := watchDir
	if w.Directory != "" {
		r := regexp.MustCompile(`^/`)
		if r.MatchString(w.Directory) {
			targetDir = w.Directory
		} else {
			targetDir = watchDir + "/" + w.Directory
		}
	}
	w.readDir(targetDir, true)
	for {
		time.Sleep(time.Duration(interval) * time.Millisecond)
		w.readDir(targetDir, false)
	}
}

func (w *Watcher) readDir(dirname string, init bool) error {
	fileInfos, err := ioutil.ReadDir(dirname)
	if err != nil {
		return err
	}
	for _, fileInfo := range fileInfos {
		name := fileInfo.Name()
		modTime := fileInfo.ModTime().UnixNano()
		switch {
		case w.exclude(name):
		case strings.HasPrefix(name, "."):
		case fileInfo.IsDir():
			if err := w.readDir(dirname+"/"+name, init); err != nil {
				return err
			}
		default:
			_, prs := w.Targets[dirname]
			if !prs {
				w.Targets[dirname] = make(map[string]FileStatus)
			}
			if init {
				w.Targets[dirname][name] = FileStatus{ModTime: modTime}
			} else {
				preservedFileInfo, prs := w.Targets[dirname][name]
				if !prs {
					w.Targets[dirname][name] = FileStatus{ModTime: modTime, LastModTime: modTime}
					w.sendJob(dirname, name, "created")
				} else if preservedFileInfo.LastModTime != 0 {
					w.Targets[dirname][name] = FileStatus{ModTime: modTime, LastModTime: 0}
				} else if preservedFileInfo.LastModTime == 0 && preservedFileInfo.ModTime != modTime {
					w.Targets[dirname][name] = FileStatus{ModTime: modTime, LastModTime: modTime}
					w.sendJob(dirname, name, "updated")
				}
			}
		}
	}
	if !init {
		preservedFileInfos, prs := w.Targets[dirname]
		if prs {
			for name, _ := range preservedFileInfos {
				exist := false
				for _, fileInfo := range fileInfos {
					if name == fileInfo.Name() {
						exist = true
						break
					}
				}
				if !exist {
					delete(w.Targets[dirname], name)
					w.sendJob(dirname, name, "deleted")
				}
			}
		}
	}
	return nil
}

func (w *Watcher) sendJob(dirname, name, action string) {
	w.JobsC <- Job{Kind: action, Path: dirname + "/" + name}
}

func (w *Watcher) exclude(filename string) bool {
	for _, excludeFilename := range w.Excludes {
		if filename == excludeFilename {
			return true
		}
	}
	return false
}
