package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kdiab/base3"
	"golang.org/x/term"
)

type GameState struct {
	entities     []entity
	wave         int
	termX        int
	termY        int
	waveComplete bool
}

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
	if e1 <= 0 {
		return entities
	}
	gap := s.x / e1
	for i := 0; i < e1; i++ {
		s.x = gap + s.width*i*2
		entities = append(entities, s)
	}

	return entities
}

func drawEntities(state *GameState, player *entity) {
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l\x1b[1;1r")
	fmt.Print(state.wave)
	if state.waveComplete == false {
		newWave(state)
	}
	for _, e := range state.entities {
		drawShape(&e)
	}
	drawShape(player)
}

func readInput(userInput chan byte) {
	var b [1]byte
	for {
		n, err := os.Stdin.Read(b[:])
		if err != nil {
			close(userInput)
			return
		}
		if n > 0 {
			userInput <- b[0]
		}
	}
}

func processInput(userInput chan byte, exitChan chan bool, state *GameState, player *entity) {
	select {
	case b, ok := <-userInput:
		if !ok {
			return
		}
		if b == 'a' {
			player.move(player, -1, 0)
		}
		if b == 'd' {
			player.move(player, 1, 0)
		}
		if b == 'n' {
			updateGame(state)
		}
		if b == 'q' || b == 3 {
			exitChan <- true
		}
	default:
	}
}

func exitGame(exitChan chan bool) {
	select {
	case <-exitChan:
		disableRawMode()
		fmt.Println("Thank you for playing!")
		os.Exit(0)
	case <-signalChan():
		disableRawMode()
		fmt.Println("Thank you for playing!")
		os.Exit(0)
	}
}

func MakeEnemies(state *GameState) (enemies []int) {
	var out []int
	base3String := base3.IntToBase3(state.wave, 4)
	for _, e := range base3String {
		out = append(out, int(e)-48)
	}
	return out
}

func newWave(state *GameState) {

	var empty []entity
	state.entities = empty
	x := state.termX
	y := state.termY

	enemies := MakeEnemies(state)

	ufo := entity{
		shape: []int{
			0b0001000,
			0b0111110,
			0b1010101,
		},
		width: 7,
		x:     x / 2,
		y:     4,
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
		x:     x / 2,
		y:     y / 2,
	}

	state.entities = append(state.entities, generateEntities(ufo, enemies[3])...)
	state.entities = append(state.entities, generateEntities(octopus, enemies[2])...)
}

func signalChan() chan os.Signal {
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGINT, syscall.SIGTERM)
	return sigTerm
}

func updateGame(state *GameState) {
	state.wave += 1
	state.waveComplete = false
}

func main() {
	enableRawMode()
	defer disableRawMode()

	x, y, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		die("Could not get terminal size: " + err.Error())
	}

	player := entity{
		shape: []int{
			0b000010000,
			0b010111010,
			0b111101111,
		},
		width: 9,
		x:     x / 2,
		y:     y - 3,
		move: func(e *entity, dx int, dy int) {
			e.x += dx
			e.y += dy
		},
	}

	state := GameState{
		wave:         1,
		termX:        x,
		termY:        y,
		waveComplete: false,
	}

	exitChan := make(chan bool)
	userInput := make(chan byte)

	go exitGame(exitChan)
	go readInput(userInput)

	for {
		processInput(userInput, exitChan, &state, &player)
		drawEntities(&state, &player)
		time.Sleep(33 * time.Millisecond)
	}
}
