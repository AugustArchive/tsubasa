// 🐇 tsubasa: Microservice to define a schema and execute it in a fast environment.
// Copyright 2022 Noel <cutie@floofy.dev>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strings"
)

type Formatter struct {
	// DisableColors is when we need to disable colour output.
	//
	// This can be overrided using the `TSUBASA_DISABLE_COLORS` environment
	// variable.
	DisableColors bool
}

var format = "Jan 02, 2006 - 15:04:05 MST"

// NewFormatter creates a new Formatter instance.
func NewFormatter() *Formatter {
	var disabledColors = false
	if os.Getenv("TSUBASA_DISABLE_COLORS") != "" {
		disabledColors = true
	}

	f := &Formatter{
		DisableColors: disabledColors,
	}

	return f
}

// Format renders a single log entry for logrus.
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	fields := make(logrus.Fields)
	for k, v := range entry.Data {
		fields[k] = v
	}

	level := f.getColourForLevel(entry.Level)
	b := &bytes.Buffer{}

	if f.DisableColors {
		fmt.Fprintf(b, "[%s] ", entry.Time.Format(format))
	} else {
		fmt.Fprintf(b, "\x1b[38;2;134;134;134m[%s] \x1b[0m", entry.Time.Format(format))
	}

	l := strings.ToUpper(entry.Level.String())
	if f.DisableColors {
		b.WriteString("[" + l[:4] + "] ")
	} else {
		b.WriteString(level)
		b.WriteString("[" + l[:4] + "] ")
		b.WriteString("\x1b[0m")
	}

	if len(fields) != 0 {
		for f, v := range fields {
			fmt.Fprintf(b, "[%s=%v] ", f, v)
		}
	}

	if entry.HasCaller() {
		var pkg string
		if strings.HasPrefix(entry.Caller.Function, "floofy.dev/tsubasa/") {
			pkg = strings.TrimPrefix(entry.Caller.Function, "floofy.dev/tsubasa/")
		} else {
			pkg = entry.Caller.Function
		}

		// Sometimes, for middleware, the `pkg` variable will contain
		// .func{int}, so we need to remove that!
		if strings.Contains(entry.Caller.Function, ".func") {
			regex, _ := regexp.Compile("\\.(func\\d+)")
			pkg = regex.ReplaceAllString(entry.Caller.Function, "")
		}

		// To preserve actual readability, the path that it is executing
		// is just :gone:!
		cwd, _ := os.Getwd()
		file := strings.TrimPrefix(strings.Replace(entry.Caller.File, cwd, "", -1), "/")

		if f.DisableColors {
			fmt.Fprintf(b, "[%s (%s:%d)] ", pkg, entry.Caller.File, entry.Caller.Line)
		} else {
			fmt.Fprintf(b, "\x1b[38;2;134;134;134m[%s (%s:%d)]\x1b[0m ", pkg, file, entry.Caller.Line)
		}
	}

	b.WriteString(strings.TrimSpace(entry.Message))
	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (f *Formatter) getColourForLevel(level logrus.Level) string {
	if f.DisableColors {
		return ""
	}

	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		// #A3B68A
		return "\x1b[1m\x1b[38;2;163;182;138m"

	case logrus.ErrorLevel, logrus.FatalLevel:
		// #994B68
		return "\x1b[1m\x1b[38;2;153;75;104m"

	case logrus.WarnLevel:
		// #F3F386
		return "\x1b[1m\x1b[38;2;243;243;134m"

	case logrus.InfoLevel:
		// #B29DF3
		return "\x1b[1m\x1b[38;2;178;157;243m"

	default:
		// #2f2f2f
		return "\x1b[1m\x1b[38;2;47;47;47m"
	}
}
