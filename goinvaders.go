package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kdiab/base3"
	"golang.org/x/term"
)

type GameState struct {
	entities     []*entity
	bullets      []*bullet
	wave         int
	termX        int
	termY        int
	waveComplete bool
}

type bullet struct {
	shape    []int
	width    int
	height   int
	x        int
	y        int
	velocity int
	damage   int
}

type entity struct {
	width    int
	y        int // line position in terminal
	x        int // column position in terminal
	shape    []int
	move     func(e *entity, dx int, dy int)
	shoot    func(e *entity) *bullet
	health   int
	damaged  bool
	alive    bool
	collided bool
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
	if s.damaged == true {
		for i, w := range shape {
			fmt.Printf("\033[38;2;245;0;0m\x1b[%d;%dH%s", s.y+i, s.x, w)
		}
	} else {
		for i, w := range shape {
			fmt.Printf("\x1b[0m\x1b[%d;%dH%s", s.y+i, s.x, w)
		}
	}
}

func generateEntities(s entity, e1 int, termX int) []*entity {
	var entities []*entity
	if e1 <= 0 {
		return entities
	}
	for i := 0; i < e1; i++ {
		r := rand.Intn(termX)
		if r < s.width {
			r += s.width
		}
		if r > termX-s.width {
			r -= s.width
		}
		newEntity := s
		newEntity.x = r
		entities = append(entities, &newEntity)
	}
	return entities
}

func drawEntities(state *GameState, player *entity) {
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l\x1b[1;1r")
	fmt.Printf("DEBUG INFO\r\nWave: %d\r\nWave in Base3: %s\r\nTerminal Width: %d\r\nPlayer Position: %d\r\nLeft Wall Collision: %t\r\nRight Wall Collision: %t\r\n", state.wave, base3.IntToBase3(state.wave, 5), state.termX, player.x, detectBoundaryCollision('l', state.termX-player.width, player.x), detectBoundaryCollision('r', state.termX-player.width, player.x))
	for _, e := range state.entities {
		fmt.Printf("Entity X: %d\r\nEntity Y: %d\r\n", e.x, e.y)
	}
	//	for _, e := range state.bullets {
	//		fmt.Printf("Bullet X: %d\r\nBullet Y: %d\r\n", e.x, e.y)
	//	}
	if state.waveComplete == true {
		newWave(state)
	}
	for i, e := range state.entities {
		if e.health <= 0 {
			state.entities[i].alive = false
		}
		dy := 0
		if detectBoundaryCollision('l', state.termX-e.width, e.x) {
			state.entities[i].collided = true
			dy = 3
		} else if detectBoundaryCollision('r', state.termX-e.width, e.x) {
			state.entities[i].collided = false
			dy = 3
		}
		dx := 0
		if e.collided == true {
			dx = 1
		} else {
			dx = -1
		}
		if e.alive {
			state.entities[i].move(state.entities[i], dx, dy)
			drawShape(e)
			state.entities[i].damaged = false
		}
	}
	drawShape(player)
	for _, b := range state.bullets {
		if b.y < 0+b.velocity || detectCollision(state, b) {
			new_bullets := removeBullet(state.bullets, b)
			state.bullets = new_bullets
		}
		drawBullet(b)
	}
	if allEnemiesKilled(state) {
		state.waveComplete = true
		updateGame(state)
	}
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
			if !detectBoundaryCollision('l', state.termX-player.width, player.x) {
				player.move(player, -2, 0)
			}
		}
		if b == 'd' {
			if !detectBoundaryCollision('r', state.termX-player.width, player.x) {
				player.move(player, 2, 0)
			}
		}
		if b == 'w' {
			bullet := player.shoot(player)
			spawnBullet(bullet, state)
		}
		if b == 'n' {
			state.waveComplete = true
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

func allEnemiesKilled(state *GameState) bool {
	for _, e := range state.entities {
		if e.alive {
			return false
		}
	}
	return true
}

func newWave(state *GameState) {

	var empty []*entity
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
		width:  7,
		x:      x - 7,
		y:      4,
		health: 5,
		alive:  true,
		move: func(e *entity, dx, dy int) {
			e.x += dx
			e.y += dy
		},
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
		x:      x - 10,
		y:      y / 2,
		health: 15,
		alive:  true,
		move: func(e *entity, dx, dy int) {

		},
	}

	state.entities = append(state.entities, generateEntities(ufo, enemies[3]+enemies[2], state.termX)...)
	state.entities = append(state.entities, generateEntities(octopus, enemies[1], state.termX)...)
	state.waveComplete = false
}

func signalChan() chan os.Signal {
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGINT, syscall.SIGTERM)
	return sigTerm
}

func updateGame(state *GameState) {
	state.wave += 1
}

func detectBoundaryCollision(direction rune, boundary int, pos int) bool {
	switch direction {
	case 'l':
		return pos < 2
	case 'r':
		return pos > boundary
	default:
		return false
	}
}

func drawBullet(b *bullet) {
	var shape []string

	for _, line := range b.shape {
		binaryString := fmt.Sprintf("%0*b", b.width, line)
		lineStr := strings.ReplaceAll(binaryString, "1", "█")
		lineStr = strings.ReplaceAll(lineStr, "0", " ")
		shape = append(shape, lineStr)
	}

	for i, w := range shape {
		fmt.Printf("\x1b[%d;%dH%s", b.y+i, b.x, w)
	}
	b.y -= b.velocity
}

func spawnBullet(bullets *bullet, state *GameState) {
	var ret []*bullet
	for i, b := range bullets.shape {
		binaryString := fmt.Sprintf("%0*b", bullets.width, b)
		for j, s := range binaryString {
			if s == '1' {
				singleBullet := bullet{
					shape:    []int{1},
					x:        bullets.x + j,
					y:        bullets.y + i,
					width:    1,
					velocity: bullets.velocity,
					damage:   bullets.damage,
				}
				ret = append(ret, &singleBullet)
			}
		}
	}
	state.bullets = append(state.bullets, ret...)
}

func removeBullet(bullets []*bullet, bulletToRemove *bullet) []*bullet {
	newBullets := []*bullet{}
	for _, bullet := range bullets {
		if bullet != bulletToRemove {
			newBullets = append(newBullets, bullet)
		}
	}
	return newBullets
}

func detectCollision(state *GameState, b *bullet) bool {
	for i, e := range state.entities {
		if (b.x >= e.x && b.x <= e.x+e.width-1) && b.y == e.y+len(e.shape) {
			state.entities[i].health -= b.damage
			state.entities[i].damaged = true
			if state.entities[i].alive {
				return true
			}
		}
	}
	return false
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
			0b100111001,
			0b111101111,
		},
		width: 9,
		x:     x / 2,
		y:     y - 3,
		move: func(e *entity, dx int, dy int) {
			e.x += dx
			e.y += dy
		},
		shoot: func(e *entity) *bullet {
			return &bullet{
				shape: []int{
					0b000010000,
					0b100010001,
				},
				x:        e.x,
				y:        e.y,
				width:    9,
				height:   2,
				velocity: 1,
				damage:   1,
			}
		},
		health: 100,
		alive:  true,
	}

	state := GameState{
		wave:         1,
		termX:        x,
		termY:        y,
		waveComplete: false,
	}
	newWave(&state)

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
