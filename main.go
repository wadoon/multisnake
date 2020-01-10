package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"io/ioutil"
	"math/rand"
	"os"
)
import "encoding/json"

const VERSION = "0.1"
const PROGRAM_NAME = "MultiSnake"

type PlayerConfig struct {
	Color    uint32
	Name     string
	KeyUp    string
	KeyDown  string
	KeyLeft  string
	KeyRight string
}

type GameConfig struct {
	Width           int32
	Height          int32
	Players         []PlayerConfig
	BackgroundColor uint32
	Food            uint32
	FieldSize       int32
	FoodColor       uint32
}

type Point struct {
	x, y int32
}

type Snake struct {
	parts                             []Point
	direction                         Point
	color                             uint32
	score                             uint
	name                              string
	keyUp, keyDown, keyLeft, keyRight sdl.Keycode
}

var config GameConfig
var players []Snake
var food []Point

func readConfig() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		panic(err)
	}
}

func initGame() {
	players = make([]Snake, len(config.Players))
	for i := 0; i < len(config.Players); i++ {
		pc := config.Players[i]
		players[i] = Snake{
			color:     pc.Color,
			name:      pc.Name,
			keyDown:   sdl.GetKeyFromName(pc.KeyDown),
			keyUp:     sdl.GetKeyFromName(pc.KeyUp),
			keyLeft:   sdl.GetKeyFromName(pc.KeyLeft),
			keyRight:  sdl.GetKeyFromName(pc.KeyRight),
			direction: Point{config.FieldSize, 0},
		}
		players[i].parts = make([]Point, 5)
		for j := 0; j < 5; j++ {
			if j == 0 {
				players[i].parts[0] = randomPoint()
			} else {
				players[i].parts[j].x = players[i].parts[j-1].x + players[i].direction.x
				players[i].parts[j].y = players[i].parts[j-1].y + players[i].direction.y
			}
		}
	}
	food = make([]Point, config.Food)
	for i := uint32(0); i < config.Food; i++ {
		food[i] = randomPoint()
	}
}

func randomRasterPoint(max int32) int32 {
	return (rand.Int31n(max) / config.FieldSize) * config.FieldSize
}

func randomPoint() Point {
	return Point{
		x: randomRasterPoint(config.Width),
		y: randomRasterPoint(config.Height),
	}
}

type Game struct {
	surface     *sdl.Surface
	window      *sdl.Window
	gameRunning bool
	font        *ttf.Font
}

func main() {
	readConfig()
	var g Game
	g.run()
}

func (g *Game) run() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()

	w, err := sdl.CreateWindow(PROGRAM_NAME+" "+VERSION,
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		config.Width, config.Height, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer w.Destroy()
	g.window = w

	// Load the font for our text
	if g.font, err = ttf.OpenFont("/usr/share/fonts/bitstream-vera/Vera.ttf", 16); err != nil {
		panic(err)
	}
	defer g.font.Close()

	s, err := g.window.GetSurface()
	g.surface = s
	if err != nil {
		panic(err)
	}

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			case *sdl.KeyboardEvent:
				if !g.gameRunning {
					println("Game starts!")
					g.gameRunning = true
					initGame()
					continue
				} else {
					g.handlePlayerKey(t.Keysym.Sym)
				}
				break
			}
		}

		if g.gameRunning {
			g.surface.FillRect(nil, config.BackgroundColor)
			for index, _ := range players {
				g.forward(&players[index], index)
			}
		} else {
			g.drawMessage("Pressed a key to start")
		}
		for idx := range players {
			g.drawPlayer(&players[idx])
		}
		g.drawFood()

		g.window.UpdateSurface()
		sdl.Delay(100)
	}
}

func (g *Game) drawMessage(content string) {
	text, err := g.font.RenderUTF8Blended(content, sdl.Color{R: 255, G: 0, B: 0, A: 255})
	defer text.Free()
	if err == nil {
		// Draw the text around the center of the window
		_ = text.Blit(nil, g.surface,
			&sdl.Rect{X: 400 - (text.W / 2), Y: 300 - (text.H / 2),
				W: 0, H: 0})
	}
}

func (g *Game) handlePlayerKey(sym sdl.Keycode) {
	for idx := range players {
		if sym == players[idx].keyUp {
			players[idx].direction = Point{0, -config.FieldSize}
		}

		if sym == players[idx].keyDown {
			players[idx].direction = Point{0, config.FieldSize}
		}

		if sym == players[idx].keyLeft {
			players[idx].direction = Point{-config.FieldSize, 0}
		}

		if sym == players[idx].keyRight {
			players[idx].direction = Point{config.FieldSize, 0}
		}
	}
}

func (g *Game) drawPlayer(current *Snake) {
	for _, part := range current.parts {
		g.drawRect(part, current.color)
	}
}

func (g *Game) forward(current *Snake, index int) {
	head := current.parts[len(current.parts)-1]
	newHead := Point{x: head.x + current.direction.x,
		y: head.y + current.direction.y,
	}
	current.parts = append(current.parts, newHead)

	foodHit := false
	for idx, f := range food {
		if newHead.x == f.x && newHead.y == f.y {
			foodHit = true
			food[idx] = randomPoint()
			current.score += 10
		}
	}
	if !foodHit {
		current.parts = current.parts[1:]
	}

	if newHead.x < 0 || newHead.y < 0 || newHead.x > config.Width || newHead.y > config.Height {
		g.drawMessage(fmt.Sprintf("Player %s lost\n", current.name))
		g.gameRunning = false
	}

	//collision with other player
	for idx, other := range players {
		for pidx, otherPart := range other.parts {
			if idx != index || pidx != len(current.parts)-1 {
				if otherPart == newHead {
					g.drawMessage(fmt.Sprintf("Player %s lost\n", current.name))
					g.gameRunning = false
				}
			}
		}
	}
}

func (g *Game) drawFood() {
	for _, f := range food {
		g.drawRect(f, config.FoodColor)
	}
}

func (g *Game) drawRect(p Point, color uint32) {
	rect := sdl.Rect{p.x, p.y, config.FieldSize, config.FieldSize}
	err := g.surface.FillRect(&rect, color)
	if err != nil {
		panic(err)
	}
}
