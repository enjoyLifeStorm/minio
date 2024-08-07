// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package logger

import (
	"context"
	"sync"
	"time"
)

// Holds a map of recently logged errors.
type logOnceType struct {
	IDMap map[interface{}]error
	sync.Mutex
}

func (l *logOnceType) logOnceConsoleIf(ctx context.Context, err error, id interface{}, errKind ...interface{}) {
	if err == nil {
		return
	}
	l.Lock()
	shouldLog := false
	prevErr := l.IDMap[id]
	if prevErr == nil {
		l.IDMap[id] = err
		shouldLog = true
	} else if prevErr.Error() != err.Error() {
		l.IDMap[id] = err
		shouldLog = true
	}
	l.Unlock()

	if shouldLog {
		consoleLogIf(ctx, err, errKind...)
	}
}

// One log message per error.
func (l *logOnceType) logOnceIf(ctx context.Context, err error, id interface{}, errKind ...interface{}) {
	if err == nil {
		return
	}
	l.Lock()
	shouldLog := false
	prevErr := l.IDMap[id]
	if prevErr == nil {
		l.IDMap[id] = err
		shouldLog = true
	} else if prevErr.Error() != err.Error() {
		l.IDMap[id] = err
		shouldLog = true
	}
	l.Unlock()

	if shouldLog {
		LogIf(ctx, err, errKind...)
	}
}

// Cleanup the map every 30 minutes so that the log message is printed again for the user to notice.
func (l *logOnceType) cleanupRoutine() {
	for {
		l.Lock()
		l.IDMap = make(map[interface{}]error)
		l.Unlock()

		time.Sleep(30 * time.Minute)
	}
}

// Returns logOnceType
func newLogOnceType() *logOnceType {
	l := &logOnceType{IDMap: make(map[interface{}]error)}
	go l.cleanupRoutine()
	return l
}

var logOnce = newLogOnceType()

// LogOnceIf - Logs notification errors - once per error.
// id is a unique identifier for related log messages, refer to cmd/notification.go
// on how it is used.
func LogOnceIf(ctx context.Context, err error, id interface{}, errKind ...interface{}) {
	if logIgnoreError(err) {
		return
	}
	logOnce.logOnceIf(ctx, err, id, errKind...)
}

// LogOnceConsoleIf - similar to LogOnceIf but exclusively only logs to console target.
func LogOnceConsoleIf(ctx context.Context, err error, id interface{}, errKind ...interface{}) {
	if logIgnoreError(err) {
		return
	}
	logOnce.logOnceConsoleIf(ctx, err, id, errKind...)
}
