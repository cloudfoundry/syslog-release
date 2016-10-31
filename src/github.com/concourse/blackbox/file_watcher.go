package blackbox

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/concourse/blackbox/syslog"
	"github.com/tedsuo/ifrit/grouper"
)

const POLL_INTERVAL = 5 * time.Second

type fileWatcher struct {
	logger *log.Logger

	sourceDir          string
	dynamicGroupClient grouper.DynamicClient

	drainerFactory syslog.DrainerFactory
}

func NewFileWatcher(
	logger *log.Logger,
	sourceDir string,
	dynamicGroupClient grouper.DynamicClient,
	drainerFactory syslog.DrainerFactory,
) *fileWatcher {
	return &fileWatcher{
		logger:             logger,
		sourceDir:          sourceDir,
		dynamicGroupClient: dynamicGroupClient,
		drainerFactory:     drainerFactory,
	}
}

func (f *fileWatcher) Watch() {
	for {
		logDirs, err := ioutil.ReadDir(f.sourceDir)
		if err != nil {
			f.logger.Fatalf("could not list directories in source dir: %s\n", err)
		}

		for _, logDir := range logDirs {
			tag := logDir.Name()
			tagDirPath := filepath.Join(f.sourceDir, tag)

			fileInfo, err := os.Stat(tagDirPath)
			if err != nil {
				f.logger.Fatalf("failed to determine if path is directory: %s\n", err)
			}

			if !fileInfo.IsDir() {
				continue
			}

			logFiles, err := ioutil.ReadDir(tagDirPath)
			if err != nil {
				f.logger.Fatalf("could not list files in log dir %s: %s\n", tag, err)
			}

			for _, logFile := range logFiles {
				if !strings.HasSuffix(logFile.Name(), ".log") {
					continue
				}

				logFileFullPath := filepath.Join(tagDirPath, logFile.Name())
				if _, found := f.dynamicGroupClient.Get(logFileFullPath); !found {
					f.dynamicGroupClient.Inserter() <- f.memberForFile(logFileFullPath)
				}
			}
		}

		time.Sleep(POLL_INTERVAL)
	}
}

func (f *fileWatcher) memberForFile(logfilePath string) grouper.Member {
	drainer, err := f.drainerFactory.NewDrainer()
	if err != nil {
		f.logger.Fatalf("could not drain to syslog: %s\n", err)
	}

	logfileDir := filepath.Dir(logfilePath)

	tag, err := filepath.Rel(f.sourceDir, logfileDir)
	if err != nil {
		f.logger.Fatalf("could not compute tag from file path %s: %s\n", logfilePath, err)
	}

	tailer := &Tailer{
		Path:    logfilePath,
		Tag:     tag,
		Drainer: drainer,
	}

	return grouper.Member{tailer.Path, tailer}
}
