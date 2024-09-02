package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type entity struct {
	width int
	y     int // line position in terminal
	x     int // column position in terminal
	shape []int
	move  func(e *entity, dx int, dy int)
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
		fmt.Printf("\x1b[%d;%dH%s", s.y+i, s.x, w)
	}
}

func generateEntities(s entity, e1 int) []entity {
	var entities []entity
	gap := s.x / e1
	for i := 0; i < e1; i++ {
		s.x = gap + s.width*i*2
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
		width: 7,
		x:     column,
		y:     1,
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
		width: 10,
		x:     column / 2,
		y:     line / 2,
	}
	spaceship := entity{
		shape: []int{
			0b000010000,
			0b010111010,
			0b111101111,
		},
		width: 9,
		x:     column / 2,
		y:     line - 3,
		move: func(e *entity, dx int, dy int) {
			e.x += dx
			e.y += dy
		},
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
			entities[0].move(&entities[0], -1, 0)
		}
		if b[0] == 'd' {
			entities[0].move(&entities[0], 1, 0)
		}
		if b[0] == 'q' {
			break
		}
		drawEntities(entities)
	}
}
