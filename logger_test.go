package lgr

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoggerNoDbg(t *testing.T) {
	tbl := []struct {
		format     string
		args       []interface{}
		rout, rerr string
	}{
		{"", []interface{}{}, "2018/01/07 13:02:34.000 INFO  \n", ""},
		{"DEBUG something 123 %s", []interface{}{"aaa"}, "", ""},
		{"[DEBUG] something 123 %s", []interface{}{"aaa"}, "", ""},
		{"INFO something 123 %s", []interface{}{"aaa"}, "2018/01/07 13:02:34.000 INFO  something 123 aaa\n", ""},
		{"[INFO] something 123 %s", []interface{}{"aaa"}, "2018/01/07 13:02:34.000 INFO  something 123 aaa\n", ""},
		{"blah something 123 %s", []interface{}{"aaa"}, "2018/01/07 13:02:34.000 INFO  blah something 123 aaa\n", ""},
		{"WARN something 123 %s", []interface{}{"aaa"}, "2018/01/07 13:02:34.000 WARN  something 123 aaa\n", ""},
		{"ERROR something 123 %s", []interface{}{"aaa"}, "2018/01/07 13:02:34.000 ERROR something 123 aaa\n",
			"2018/01/07 13:02:34.000 ERROR something 123 aaa\n"},
	}
	rout, rerr := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	l := New(Out(rout), Err(rerr), Msec)
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }

	for i, tt := range tbl {
		rout.Reset()
		rerr.Reset()
		t.Run(fmt.Sprintf("check-%d", i), func(t *testing.T) {
			l.Logf(tt.format, tt.args...)
			assert.Equal(t, tt.rout, rout.String())
			assert.Equal(t, tt.rerr, rerr.String())
		})
	}
}

func TestLoggerWithDbg(t *testing.T) {
	tbl := []struct {
		format     string
		args       []interface{}
		rout, rerr string
	}{
		{"", []interface{}{},
			"2018/01/07 13:02:34.123 INFO  {lgr/logger_test.go:79 lgr.TestLoggerWithDbg.func2} \n", ""},
		{"DEBUG something 123 %s", []interface{}{"aaa"},
			"2018/01/07 13:02:34.123 DEBUG {lgr/logger_test.go:79 lgr.TestLoggerWithDbg.func2} something 123 aaa\n", ""},
		{"[DEBUG] something 123 %s", []interface{}{"aaa"},
			"2018/01/07 13:02:34.123 DEBUG {lgr/logger_test.go:79 lgr.TestLoggerWithDbg.func2} something 123 aaa\n", ""},
		{"INFO something 123 %s", []interface{}{"aaa"},
			"2018/01/07 13:02:34.123 INFO  {lgr/logger_test.go:79 lgr.TestLoggerWithDbg.func2} something 123 aaa\n", ""},
		{"[INFO] something 123 %s", []interface{}{"aaa"},
			"2018/01/07 13:02:34.123 INFO  {lgr/logger_test.go:79 lgr.TestLoggerWithDbg.func2} something 123 aaa\n", ""},
		{"blah something 123 %s", []interface{}{"aaa"},
			"2018/01/07 13:02:34.123 INFO  {lgr/logger_test.go:79 lgr.TestLoggerWithDbg.func2} blah something 123 aaa\n", ""},
		{"WARN something 123 %s", []interface{}{"aaa"},
			"2018/01/07 13:02:34.123 WARN  {lgr/logger_test.go:79 lgr.TestLoggerWithDbg.func2} something 123 aaa\n", ""},
		{"ERROR something 123 %s", []interface{}{"aaa"},
			"2018/01/07 13:02:34.123 ERROR {lgr/logger_test.go:79 lgr.TestLoggerWithDbg.func2} something 123 aaa\n",
			"2018/01/07 13:02:34.123 ERROR {lgr/logger_test.go:79 lgr.TestLoggerWithDbg.func2} something 123 aaa\n"},
	}

	rout, rerr := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	l := New(Debug, CallerFile, CallerFunc, Out(rout), Err(rerr), Msec)
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 123000000, time.Local) }

	for i, tt := range tbl {
		rout.Reset()
		rerr.Reset()
		t.Run(fmt.Sprintf("check-%d", i), func(t *testing.T) {
			l.Logf(tt.format, tt.args...)
			assert.Equal(t, tt.rout, rout.String())
			assert.Equal(t, tt.rerr, rerr.String())
		})
	}

	l = New(Debug, Out(rout), Err(rerr), Msec) // no caller
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }
	rout.Reset()
	rerr.Reset()
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.000 DEBUG something 123 err\n", rout.String())
	assert.Equal(t, "", rerr.String())

	l = New(Debug, Out(rout), Err(rerr), CallerFile, Msec) // caller file only
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }
	rout.Reset()
	rerr.Reset()
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.000 DEBUG {lgr/logger_test.go:97} something 123 err\n", rout.String())

	l = New(Debug, Out(rout), Err(rerr), CallerFunc, Msec) // caller func only
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }
	rout.Reset()
	rerr.Reset()
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.000 DEBUG {lgr.TestLoggerWithDbg} something 123 err\n", rout.String())

	l = New(Debug, Out(rout), Err(rerr), CallerFunc) // caller func only, no msec
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }
	rout.Reset()
	rerr.Reset()
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34 DEBUG {lgr.TestLoggerWithDbg} something 123 err\n", rout.String())
}

func TestLoggerWithPkg(t *testing.T) {
	rout, rerr := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	l := New(Debug, Out(rout), Err(rerr), CallerPkg, Msec)
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 123000000, time.Local) }
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.123 DEBUG {lgr} something 123 err\n", rout.String())

	l = New(Debug, Out(rout), Err(rerr), CallerPkg, CallerFile, Msec)
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 123000000, time.Local) }
	rout.Reset()
	rerr.Reset()
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.123 DEBUG {lgr/logger_test.go:126} something 123 err\n", rout.String())

	l = New(Debug, Out(rout), Err(rerr), CallerPkg, CallerFunc, Msec)
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 123000000, time.Local) }
	rout.Reset()
	rerr.Reset()
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.123 DEBUG {lgr.TestLoggerWithPkg} something 123 err\n", rout.String())
}

func TestLoggerIgnoreCallers(t *testing.T) {
	rout, rerr := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	l := New(Debug, Out(rout), Err(rerr), CallerPkg, Msec, CallerIgnore("lgr"))
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 123000000, time.Local) }
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.123 DEBUG {go-pkgz} something 123 err\n", rout.String())

	l = New(Debug, Out(rout), Err(rerr), CallerFile, CallerFunc, Msec, CallerIgnore("lgr"))
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 123000000, time.Local) }
	rout.Reset()
	rerr.Reset()
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.123 DEBUG {lgr/logger_test.go:148 lgr.TestLoggerIgnoreCallers} something 123 err\n", rout.String())
}

func TestLoggerWithLevelBraces(t *testing.T) {
	rout, rerr := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	l := New(Debug, Out(rout), Err(rerr), LevelBraces, Msec)
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 123000000, time.Local) }
	l.Logf("[DEBUG] something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.123 [DEBUG] something 123 err\n", rout.String())

	rout.Reset()
	l.Logf("something 123 %s", "err")
	assert.Equal(t, "2018/01/07 13:02:34.123 [INFO]  something 123 err\n", rout.String())
}

func TestLoggerWithPanic(t *testing.T) {
	fatalCalls := 0
	rout, rerr := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	l := New(Debug, CallerFunc, Out(rout), Err(rerr))
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }
	l.fatal = func() { fatalCalls++ }

	l.Logf("[PANIC] oh my, panic now! %v", errors.New("bad thing happened"))
	assert.Equal(t, 1, fatalCalls)
	assert.Equal(t, "2018/01/07 13:02:34 PANIC {lgr.TestLoggerWithPanic} oh my, panic now! bad thing happened\n", rout.String())

	t.Logf(rerr.String())
	assert.True(t, strings.HasPrefix(rerr.String(), "2018/01/07 13:02:34 PANIC"))
	assert.True(t, strings.Contains(rerr.String(), "github.com/go-pkgz/lgr.getDump"))
	assert.True(t, strings.Contains(rerr.String(), "go-pkgz/lgr/logger.go:"))

	rout.Reset()
	rerr.Reset()
	l.Logf("[FATAL] oh my, panic now! %v", errors.New("bad thing happened"))
	assert.Equal(t, 2, fatalCalls)
	assert.Equal(t, "2018/01/07 13:02:34 FATAL {lgr.TestLoggerWithPanic} oh my, panic now! bad thing happened\n", rout.String())

	rout.Reset()
	rerr.Reset()
	fatalCalls = 0
	l = New(Out(rout), Err(rerr))
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }
	l.fatal = func() { fatalCalls++ }
	l.Logf("[PANIC] oh my, panic now! %v", errors.New("bad thing happened"))
	assert.Equal(t, 1, fatalCalls)
	assert.Equal(t, "2018/01/07 13:02:34 PANIC oh my, panic now! bad thing happened\n", rout.String())
	assert.True(t, strings.HasPrefix(rerr.String(), "2018/01/07 13:02:34 PANIC"))
	assert.True(t, strings.Contains(rerr.String(), "github.com/go-pkgz/lgr.getDump"))
	assert.True(t, strings.Contains(rerr.String(), "go-pkgz/lgr/logger.go:"))
}

func TestLoggerConcurrent(t *testing.T) {
	rout, rerr := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	l := New(Debug, Out(rout), Err(rerr))
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }

	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func(i int) {
			l.Logf("[DEBUG] test test 123 debug message #%d, %v", i, errors.New("some error"))
			wg.Done()
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 1001, len(strings.Split(rout.String(), "\n")))
	assert.Equal(t, "", rerr.String())
}

func TestCaller(t *testing.T) {
	var l *Logger

	filePath, line, funcName := l.caller(0)
	assert.True(t, strings.HasSuffix(filePath, "go-pkgz/lgr/logger_test.go"), filePath)
	assert.Equal(t, 222, line)
	assert.Equal(t, funcName, "github.com/go-pkgz/lgr.TestCaller")

	f := func() {
		filePath, line, funcName = l.caller(1)
	}
	f()
	assert.True(t, strings.HasSuffix(filePath, "go-pkgz/lgr/logger_test.go"), filePath)
	assert.Equal(t, 230, line)
	assert.Equal(t, funcName, "github.com/go-pkgz/lgr.TestCaller")
}

func BenchmarkNoDbg(b *testing.B) {
	rout, rerr := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	l := New(Out(rout), Err(rerr))
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }

	e := errors.New("some error")
	for n := 0; n < b.N; n++ {
		l.Logf("[INFO] test test 123 debug message #%d, %v", n, e)
	}
}

func BenchmarkWithDbg(b *testing.B) {
	rout, rerr := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	l := New(Debug, CallerFile, CallerFunc, Out(rout), Err(rerr))
	l.now = func() time.Time { return time.Date(2018, 1, 7, 13, 2, 34, 0, time.Local) }

	e := errors.New("some error")
	for n := 0; n < b.N; n++ {
		l.Logf("INFO test test 123 debug message #%d, %v", n, e)
	}
}
