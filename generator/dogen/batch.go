package dogen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fioncat/go-gendb/misc/cct"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
)

type batchTask struct {
	path     string
	content  string
	confPath string
}

func work(v interface{}) error {
	task := v.(*batchTask)
	ok := One(task.path, task.confPath, task.content)
	if !ok {
		return errors.New("batch generate failed")
	}
	return nil
}

func Batch(root string, confPath string) bool {
	log.Infof(`[batch] begin to fetch from "%s"`, root)
	var tasks []*batchTask
	start := time.Now()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := filepath.Base(path)
		if strings.HasSuffix(filename, ".go") {
			task, err := isAccept(path)
			if err != nil {
				return err
			}
			if task == nil {
				return nil
			}
			task.confPath = confPath
			tasks = append(tasks, task)
			log.Infof("fetch task: %s", task.path)
		}
		return nil
	})
	if err != nil {
		fmt.Println(errors.Trace("fetch failed", err))
		return false
	}
	if len(tasks) == 0 {
		fmt.Println("tagged go file is not found, nothing to do.")
		return false
	}
	log.Infof("[batch] fetch done, found %d tasks, took: %v",
		len(tasks), time.Since(start))

	start = time.Now()
	nWorkers := runtime.NumCPU()
	log.Infof("found %d CPU, will start %d worker(s)", nWorkers, nWorkers)

	wp := cct.NewPool(len(tasks), nWorkers, work)
	wp.Start()

	for _, task := range tasks {
		wp.Do(task)
	}

	if err := wp.Wait(); err != nil {
		fmt.Println(err)
		return false
	}
	log.Infof("[batch] all task(s) done, took: %v", time.Since(start))

	return true
}

func isAccept(path string) (*batchTask, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)
	lines := strings.Split(content, "\n")

	hasTag := false
	for _, line := range lines {
		if strings.HasPrefix(line, "// +gendb") {
			hasTag = true
			break
		}
		if strings.HasPrefix(line, "package") {
			break
		}
	}
	if hasTag {
		return &batchTask{
			content: content,
			path:    path,
		}, nil
	}
	return nil, nil
}
