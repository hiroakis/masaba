package main

import (
	"os"
	"testing"

	"regexp"

	"bytes"

	"strings"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	fname := "test.txt"
	jsonStr := `{"name":"foo","value":"bar"}`
	type jsonStrSt struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	f, _ := os.Create(fname)
	defer func() {
		f.Close()
		os.Remove(fname)
	}()

	var err error

	err = run("echo", []string{jsonStr}, f)
	assert.Nil(t, err)

	// exists
	err = run("ls", []string{"-1", fname}, nil)
	assert.Nil(t, err)

	// can load into buffer
	oo := &bytes.Buffer{}
	err = run("cat", []string{fname}, oo)
	assert.Nil(t, err)
	assert.Equal(t, true, strings.Contains(oo.String(), jsonStr), "should be foo")

	// can load into interafce
	var v interface{}
	err = run("cat", []string{fname}, &v)
	assert.Nil(t, err)
	w, ok := v.(map[string]interface{})
	assert.Equal(t, true, ok, "should be true")
	assert.Equal(t, "foo", w["name"], "should be foo")
	assert.Equal(t, "bar", w["value"], "should be bar")

	// can load into a struct
	var j jsonStrSt
	err = run("cat", []string{fname}, &j)
	assert.Nil(t, err)
	assert.Equal(t, "foo", j.Name, "should be foo")
	assert.Equal(t, "bar", j.Value, "should be bar")
}

func TestNowString(t *testing.T) {
	timePattern := regexp.MustCompile(`(?m)^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}$`)
	assert.Equal(t, true, timePattern.MatchString(nowString()), "Format should be YYYY-MM-DD hh:mm:ss")
}

func TestTrafficHumanReadable(t *testing.T) {
	assert.Equal(t, "1.00B/s", trafficHumanReadable(1), "should be 1.00B/s")
	assert.Equal(t, "1.00K/s", trafficHumanReadable(1*1024), "should be 1.00K/s")
	assert.Equal(t, "4.00K/s", trafficHumanReadable(4*1024), "should be 4.00K/s")
	assert.Equal(t, "1.00M/s", trafficHumanReadable(1024*1024), "should be 1.00M/s")
}

func TestMemoryHumanReadable(t *testing.T) {
	assert.Equal(t, "1.00B", memoryHumanReadable(1), "should be 1.00B")
	assert.Equal(t, "1.00K", memoryHumanReadable(1*1024), "should be 1.00K")
	assert.Equal(t, "4.00K", memoryHumanReadable(4*1024), "should be 4.00K")
	assert.Equal(t, "1.00M", memoryHumanReadable(1024*1024), "should be 1.00M")
}

func TestCpuDigits(t *testing.T) {
	assert.Equal(t, "1.00", cpuDigits(1), "should be 1.00")
	assert.Equal(t, "50.00", cpuDigits(50), "should be 50.00")
	assert.Equal(t, "100.00", cpuDigits(100), "should be 100.00")
	assert.Equal(t, "1000.00", cpuDigits(1000), "should be 1000.00")
}
