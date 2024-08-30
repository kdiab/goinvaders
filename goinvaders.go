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
	fmt.Print("\x1b[2J\x1b[H\x1b[?25h")
}

func enableRawMode() {
	var err error
	state, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		die("Error enabling raw mode: " + err.Error())
	}
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l\x1b[1;1r")
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
		fmt.Printf("\x1b[%d;%dH%s", s.line+i, s.column, w)
	}
}

func generateEntities(s entity, e1 int) []entity {
	var entities []entity
	gap := s.column / e1
	for i := 0; i < e1; i++ {
		s.column = gap + s.width*i*2
		entities = append(entities, s)
	}

	return entities
}

func drawEntities(entities []entity) {
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l\x1b[1;1r")
	for _, e := range entities {
		drawShape(&e)
	}
}

func Move(s *entity, x int, y int) {
	s.column += x
	s.line += y
}

func main() {
	enableRawMode()
	defer disableRawMode()

	column, line, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		die("Could not get terminal size: " + err.Error())
	}

	var entities []entity

	ufo := entity{
		shape: []int{
			0b0001000,
			0b0111110,
			0b1010101,
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
	spaceship := entity{
		shape: []int{
			0b000010000,
			0b010111010,
			0b111101111,
		},
		width:  9,
		column: column / 2,
		line:   line - 3,
	}

	entities = append(entities, spaceship)
	entities = append(entities, generateEntities(ufo, 14)...)
	entities = append(entities, generateEntities(octopus, 1)...)

	drawEntities(entities)
	for {
		var b [3]byte
		if _, err := os.Stdin.Read(b[:]); err != nil {
			break
		}
		if b[0] == 'a' {
			Move(&entities[0], -1, 0)
		}
		if b[0] == 'd' {
			Move(&entities[0], 1, 0)
		}
		if b[0] == 'q' {
			break
		}
		drawEntities(entities)
	}
}
