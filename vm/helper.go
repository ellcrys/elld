package vm

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

var dockerCmd *exec.Cmd

//HasDocker checks if system has docker installed
func HasDocker() bool {

	//docker -v command
	dockerCmd = exec.Command("docker", "-v")

	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := dockerCmd.StdoutPipe()
	stderrIn, _ := dockerCmd.StderrPipe()

	//exec command
	dockerCmd.Start()

	//Capture stdout
	go func() {
		stdout, errStdout = Capture(os.Stdout, stdoutIn)
	}()

	//Capture stderr
	go func() {
		stderr, errStderr = Capture(os.Stderr, stderrIn)
	}()

	//Wait for outputs from command
	err := dockerCmd.Wait()
	if err != nil {
		return false
	}

	if errStdout != nil || errStderr != nil {
		vmLog.Fatal("failed to capture stdout or stderr\n")
	}

	outStr, errStr := string(stdout), string(stderr)

	if outStr != "" {
		return true
	}
	if errStr != "" {
		return false
	}

	return false
}

//Capture commandline outputs
func Capture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

func readerToChan(reader *bytes.Buffer, exit <-chan bool) <-chan string {
	c := make(chan string)
	go func() {

		for {
			select {
			case <-exit:
				close(c)
				return
			default:
				line, err := reader.ReadString('\n')

				if err != nil && err != io.EOF {
					close(c)
					return
				}

				line = strings.TrimSpace(line)
				if line != "" {
					c <- line
				}
			}
		}
	}()

	return c
}
