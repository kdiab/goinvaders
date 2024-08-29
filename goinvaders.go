package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type entity struct {
	width  int
	line   int // y position in terminal
	column int // x position in terminal
	shape  []int
}

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
	fmt.Print("\x1b[2J\x1b[H")
}

func enableRawMode() {
	var err error
	state, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		die("Error enabling raw mode: " + err.Error())
	}
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l")
}

func drawShape(s *entity) {
	var shape []string

	for _, line := range s.shape {
		binaryString := fmt.Sprintf("%0*b", s.width, line)
		lineStr := strings.ReplaceAll(binaryString, "1", "█")
		lineStr = strings.ReplaceAll(lineStr, "0", " ")
		shape = append(shape, lineStr)
	}

	for i, w := range shape {
		fmt.Printf("\x1b[%d;%dH", s.line+i, s.column)
		fmt.Printf("%s", w)
	}
}

func drawShapes(s *entity, amount int) {
	var shape []string
	position := s.column / amount

	for _, line := range s.shape {
		binaryString := fmt.Sprintf("%0*b", s.width, line)
		lineStr := strings.ReplaceAll(binaryString, "1", "█")
		lineStr = strings.ReplaceAll(lineStr, "0", " ")
		shape = append(shape, lineStr)
	}

	for i := range amount {
		for j, w := range shape {
			fmt.Printf("\x1b[%d;%dH", s.line+j, position*i+amount)
			fmt.Printf("%s", w)
		}
	}
}

func main() {
	enableRawMode()
	defer disableRawMode()

	column, line, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		die("Could not get terminal size: " + err.Error())
	}

	spaceship := entity{
		shape: []int{
			0b000010000,
			0b010111010,
			0b011111110,
		},
		width:  9,
		column: column / 2,
		line:   line - 3,
	}
	ufo := entity{
		shape: []int{
			0b0001000, // Row 1
			0b0111110, // Row 2
			0b1010101, // Row 3
		},
		width:  7,
		column: column,
		line:   1,
	}
	octopus := entity{
		shape: []int{
			0b0011111100,
			0b0110011010,
			0b1101111011,
			0b1101111011,
			0b0111111110,
			0b0011011000,
			0b0110011010,
			0b1100000011,
		},
		width:  10,
		column: column / 2,
		line:   line / 2,
	}

	drawShape(&spaceship)
	drawShape(&octopus)
	drawShapes(&ufo, 8)

	for {
		var b [3]byte
		if _, err := os.Stdin.Read(b[:]); err != nil {
			break
		}
		if b[0] == 'q' {
			break
		}
	}
}
