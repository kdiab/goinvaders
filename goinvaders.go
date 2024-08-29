package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var state *term.State

func die(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func disableRawMode() {
	err := term.Restore(int(os.Stdin.Fd()), state)
	if err != nil {
		die("Could not restore terminal state: " + err.Error())
	}
}

func enableRawMode() {
	fmt.Print("\033[2J\033H")
	var err error
	state, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		die("Error enabling raw mode: " + err.Error())
	}
}

func main() {
	enableRawMode()
	defer disableRawMode()

	for {
		var b [1]byte
		if _, err := os.Stdin.Read(b[:]); err != nil {
			break
		}
		if b[0] == 'q' {
			break
		}
		fmt.Printf("%c", b[0])
	}
}
