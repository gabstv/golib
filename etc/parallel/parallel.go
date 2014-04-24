package parallel

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"
)

var (
	commands []*exec.Cmd
	stop1    bool
)

func Run(programs ...string) {

	commands = make([]*exec.Cmd, len(programs))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		for _, v := range commands {
			if v != nil {
				if v.Process != nil {
					v.Process.Signal(os.Interrupt)
				}
			}
		}
		stop1 = true
	}()

	for k, v := range programs {
		fmt.Println("STARTING ~~~ " + v)
		commands[k] = runProgram(v)
		time.Sleep(time.Millisecond * 500)
	}

	for !stop1 {
		time.Sleep(time.Second)
	}
}

func runProgram(raw string) *exec.Cmd {
	cmd0 := parseStrings(raw)
	cmd1 := exec.Command(cmd0[0], cmd0[1:]...)
	cmd1.Stderr = os.Stderr
	cmd1.Stdout = os.Stdout
	err := cmd1.Start()
	if err != nil {
		fmt.Println(err.Error())
	}
	return cmd1
}

func parseStrings(raw string) []string {
	var buffer bytes.Buffer
	outp := make([]string, 0)
	raw = strings.TrimSpace(raw)
	insideQ := false
	lastQ := false
	for _, v := range raw {
		if v == '\\' {
			if lastQ {
				buffer.WriteRune(v)
				lastQ = false
			} else {
				lastQ = true
			}
		} else {
			if v == '"' {
				if !insideQ && !lastQ {
					insideQ = true
				} else if insideQ && !lastQ {
					insideQ = false
					outp = append(outp, buffer.String())
					buffer.Truncate(0)
				}
				if lastQ {
					buffer.WriteRune(v)
				}
			} else {
				if v == ' ' {
					if insideQ || lastQ {
						buffer.WriteRune(v)
					} else {
						outp = append(outp, buffer.String())
						buffer.Truncate(0)
					}
				} else {
					buffer.WriteRune(v)
				}
			}
			lastQ = false
		}
	}
	if buffer.Len() > 0 {
		outp = append(outp, buffer.String())
		buffer.Truncate(0)
	}
	return outp
}
