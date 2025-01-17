/*
Copyright 2023 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"go.uber.org/zap"
)

type JobLogger struct {
	mu      sync.Mutex
	writer  io.Writer
	logger  *zap.SugaredLogger
	logPath string
}

func NewJobLogger(logfile string) *JobLogger {
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		Panicf("failed to open log file: %s", err)
	}

	cfg := &Config{
		Level:      "debug",
		Filename:   logfile,
		SendToFile: true,
		NoCaller:   true,
		NoLogLevel: true,
	}

	return &JobLogger{
		mu:      sync.Mutex{},
		writer:  file,
		logger:  InitJobLogger(cfg),
		logPath: logfile,
	}
}

func (l *JobLogger) Printf(format string, a ...any) {
	if l.logger == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := fmt.Fprintf(l.writer, format, a...)
	if err != nil {
		l.logger.Errorf("Failed to write to log: %v\n", err)
	}
}

func (l *JobLogger) Println(args ...interface{}) {
	if l.logger == nil {
		return
	}

	raw := fmt.Sprintln(args...)

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := io.WriteString(l.writer, raw); err != nil {
		Errorf("Failed to write to log: %v\n", err)
	}
}

func (l *JobLogger) Debugf(args ...interface{}) {
	if l.logger == nil {
		return
	}

	raw := fmt.Sprint(args...)

	l.logger.Debug(raw)
}

func (l *JobLogger) Infof(args ...interface{}) {
	if l.logger == nil {
		return
	}

	raw := fmt.Sprint(args...)

	l.logger.Infof(raw)
}

func (l *JobLogger) Warnf(args ...interface{}) {
	if l.logger == nil {
		return
	}

	raw := fmt.Sprint(args...)

	l.logger.Warnf(raw)
}

func (l *JobLogger) Errorf(args ...interface{}) {
	if l.logger == nil {
		return
	}

	raw := fmt.Sprint(args...)

	l.logger.Errorf(raw)
}

func (l *JobLogger) Fatalf(args ...interface{}) {
	if l.logger == nil {
		return
	}

	raw := fmt.Sprint(args...)

	l.logger.Fatalf(raw)
}

func (l *JobLogger) Panicf(args ...interface{}) {
	if l.logger == nil {
		return
	}

	raw := fmt.Sprint(args...)

	l.logger.Panicf(raw)
}

func (l *JobLogger) Write(p []byte) {
	if l.logger == nil {
		return
	}

	raw := string(p)

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := io.WriteString(l.writer, raw); err != nil {
		Errorf("Failed to write to log: %v\n", err)
	}
}

// 读取文件内容到字符串
func readFileToString(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadByRowNum TODO: 需要优化，按行读取io过多效率低
func (l *JobLogger) ReadByRowNum(offset, num uint) ([]byte, uint, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var EOFErr bool

	file, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	defer file.Close()

	if err != nil {
		Panicf("failed to open log file: %s", err)
	}

	// Create a buffered reader
	reader := bufio.NewReader(file)

	// Counter to track the current line number
	lineCount := uint(0)

	// Buffer to store the read lines
	var resultBuffer bytes.Buffer

	// Read the file line by line until reaching the specified line count or end of file
	for lineCount < offset+num {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				EOFErr = true
				break // End of file
			}
			return nil, 0, EOFErr, fmt.Errorf("failed to read log line: %v", err)
		}

		// If the current line number is within the specified range, append the line data to the result buffer
		if lineCount >= offset {
			resultBuffer.WriteString(line)
		}

		lineCount++
	}

	return resultBuffer.Bytes(), lineCount, EOFErr, nil
}

func (l *JobLogger) GetLogfilePath() string {
	return l.logPath
}

func (l *JobLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	err := l.logger.Sync()
	if err != nil {
		Errorf("failed to sync job logger, error: %s", err)
	}
}
