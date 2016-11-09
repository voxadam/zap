// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zap

// TeeLogger creates a Logger that duplicates its log calls to two or
// more loggers. It is similar to io.MultiWriter.
//
// For each logging level methad (.Debug, .Info, etc), the TeeLogger calls each
// sub-logger's level method.
//
// Exceptions are made for the Fatal and Panic methods: the returned logger
// calls .Log(FatalLevel, ...) and .Log(PanicLevel, ...). Only after all
// sub-loggers have received the message, then the TeeLogger terminates the
// process (using os.Exit or panic() per usual semantics).
//
// DFatal is handled similarly to Fatal and Panic, since it is not actually a
// level; each sub-logger's DFatal method dynamically chooses to either cal
// Error or Fatal.
//
// Check returns a CheckedMessage against the TeeLogger itself if at least one
// of the sub-logger Checks returns an OK CheckedMessage.
func TeeLogger(logs ...Logger) Logger {
	switch len(logs) {
	case 0:
		return nil
	case 1:
		return logs[0]
	default:
		return multiLogger(logs)
	}
}

type multiLogger []Logger

func (ml multiLogger) Log(lvl Level, msg string, fields ...Field) {
	ml.log(lvl, msg, fields)
}

func (ml multiLogger) Debug(msg string, fields ...Field) {
	for _, log := range ml {
		log.Debug(msg, fields...)
	}
}

func (ml multiLogger) Info(msg string, fields ...Field) {
	for _, log := range ml {
		log.Info(msg, fields...)
	}
}

func (ml multiLogger) Warn(msg string, fields ...Field) {
	for _, log := range ml {
		log.Warn(msg, fields...)
	}
}

func (ml multiLogger) Error(msg string, fields ...Field) {
	for _, log := range ml {
		log.Error(msg, fields...)
	}
}

func (ml multiLogger) Panic(msg string, fields ...Field) {
	ml.log(PanicLevel, msg, fields)
	panic(msg)
}

func (ml multiLogger) Fatal(msg string, fields ...Field) {
	ml.log(FatalLevel, msg, fields)
	_exit(1)
}

func (ml multiLogger) log(lvl Level, msg string, fields []Field) {
	for _, log := range ml {
		log.Log(lvl, msg, fields...)
	}
}

func (ml multiLogger) DFatal(msg string, fields ...Field) {
	for _, log := range ml {
		log.DFatal(msg, fields...)
	}
}

func (ml multiLogger) With(fields ...Field) Logger {
	ml = append([]Logger(nil), ml...)
	for i, log := range ml {
		ml[i] = log.With(fields...)
	}
	return multiLogger(ml)
}

func (ml multiLogger) Check(lvl Level, msg string) *CheckedMessage {
lvlSwitch:
	switch lvl {
	case PanicLevel, FatalLevel:
		// Panic and Fatal should always cause a panic/exit, even if the level
		// is disabled.
		break
	default:
		for _, log := range ml {
			if cm := log.Check(lvl, msg); cm.OK() {
				break lvlSwitch
			}
		}
		return nil
	}
	return NewCheckedMessage(ml, lvl, msg)
}
