package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"image"
	_ "image/png"
	"io/ioutil"
	"math/rand"
	"os"
)
import "encoding/json"

const VERSION = "0.1"
const PROGRAM_NAME = "MultiSnake"

type PlayerConfig struct {
	Color uint32

	Name     string
	KeyUp    string
	KeyDown  string
	KeyLeft  string
	KeyRight string

	ControllerKeyUp    string
	ControllerKeyDown  string
	ControllerKeyLeft  string
	ControllerKeyRight string
}

type GameConfig struct {
	// the width in pixels of the game area
	Width int32
	// the height in pixels of the game area
	Height int32

	//Settings for each player. Length of the array determines the number of players
	Players []PlayerConfig
	//Background color of the game arena
	BackgroundColor uint32
	//Size of a cell in the raster. Used to make snake and objects bigger
	FieldSize int32
	//number of foods on the arena
	Food uint32
	//Color of the food
	FoodColor uint32
	//increment of the score, when eating a food
	FoodScore uint32
	//number of special food cells on the field
	SuperFood uint32
	//color of the special food
	SuperFoodColor uint32
	//increment of the score for eating a special food
	SuperFoodScore uint32
	//List of b/w PNG images, which describes the obstacle in the game arena
	//The of the images need to be Width/FieldSize x Height/FieldSize.
	//In this example the PNG is 80x80 pixels.
	//Obstacles are randomly chosen at game start.
	Obstacles      []string
	ObstaclesColor uint32 //Color of the obstacles
	//If a snake leaves the arena, it will appear on the opposite side
	CycleBorder bool
}

// A Point within the game arena.
type Point struct {
	x, y int32
}

//The snake object of a player.
type Snake struct {
	parts                             []Point
	direction                         Point
	color                             uint32
	score                             uint
	name                              string
	keyUp, keyDown, keyLeft, keyRight sdl.Keycode
	sdlColor                          sdl.Color
	lost                              bool
}

var config GameConfig
var players []Snake
var food []Point

func readConfig(configFile string) {
	jsonFile, err := os.Open(configFile)
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

func (g *Game) initGame() {
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
			sdlColor: sdl.Color{R: uint8(pc.Color >> 24 & 0xff),
				G: uint8(pc.Color >> 16 & 0xFF),
				B: uint8(pc.Color >> 8 & 0xFF),
				A: uint8(pc.Color & 0xFF),
			},
			lost: false,
		}
		players[i].parts = make([]Point, 5)
		for j := 0; j < 5; j++ {
			if j == 0 {
				players[i].parts[0] = g.randomPoint()
			} else {
				players[i].parts[j].x = players[i].parts[j-1].x + players[i].direction.x
				players[i].parts[j].y = players[i].parts[j-1].y + players[i].direction.y
			}
		}
	}
	food = make([]Point, config.Food)
	for i := uint32(0); i < config.Food; i++ {
		food[i] = g.randomPoint()
	}
	if len(config.Obstacles) > 0 {
		g.obstacles = readObstacles(config.Obstacles[0])
	}
}

func randomRasterPoint(max int32) int32 {
	return (rand.Int31n(max) / config.FieldSize) * config.FieldSize
}

func (g *Game) randomPoint() Point {
	p := Point{
		x: randomRasterPoint(config.Width),
		y: randomRasterPoint(config.Height),
	}
	for idx := range g.obstacles {
		if g.obstacles[idx] == p {
			return g.randomPoint()
		}
	}

	for idx := range food {
		if food[idx] == p {
			return g.randomPoint()
		}
	}

	for idx := range players {
		for pidx := range players[idx].parts {
			if players[idx].parts[pidx] == p {
				return g.randomPoint()
			}
		}
	}
	return p
}

type Game struct {
	surface     *sdl.Surface
	window      *sdl.Window
	gameRunning bool
	font        *ttf.Font
	obstacles   []Point
}

func main() {
	configFile := ""
	if len(os.Args) == 1 {
		configFile = "config.json"
	} else {
		configFile = os.Args[1]
	}
	readConfig(configFile)
	var g Game
	g.run()
}

func readObstacles(imagePath string) []Point {
	var points []Point
	infile, err := os.Open(imagePath)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	src, _, err := image.Decode(infile)
	if err != nil {
		panic(err)
	}

	maxWidth := config.Width / config.FieldSize
	maxHeight := config.Height / config.FieldSize
	for x := int32(0); x < maxWidth; x++ {
		for y := int32(0); y < maxHeight; y++ {
			r, g, b, _ := src.At(int(x), int(y)).RGBA()
			if r == 0 && g == 0 && b == 0 {
				points = append(points, Point{x * config.FieldSize, y * config.FieldSize})
			}
		}
	}
	return points
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
					g.initGame()
					continue
				} else {
					g.handlePlayerKey(t.Keysym.Sym)
				}
				break
			case *sdl.JoyHatEvent:
				switch t.Value {
				case sdl.HAT_LEFTUP:
				case sdl.HAT_UP:
				case sdl.HAT_RIGHTUP:

				case sdl.HAT_RIGHT:

				case sdl.HAT_RIGHTDOWN:

				case sdl.HAT_DOWN:

				case sdl.HAT_LEFTDOWN:

				case sdl.HAT_LEFT:

				case sdl.HAT_CENTERED:

				}
			}
		}

		if g.gameRunning {
			g.surface.FillRect(nil, config.BackgroundColor)
			for index, _ := range players {
				g.forward(&players[index], index)
			}

			for idx := range players {
				g.drawPlayer(&players[idx])
			}

			for idx := range g.obstacles {
				g.drawRect(g.obstacles[idx], config.ObstaclesColor)
			}

			g.drawFood()
			g.drawScore()
			g.decideWinner()
		} else {
			g.drawMessage("Pressed a key to start")
		}
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

func (g *Game) drawScore() {
	for i := 0; i < len(players); i++ {
		text, err := g.font.RenderUTF8Blended(fmt.Sprintf("%s %d", players[i].name, players[i].score),
			players[i].sdlColor)
		defer text.Free()
		if err == nil {
			_ = text.Blit(nil, g.surface,
				&sdl.Rect{X: int32(100 + 100*i), Y: 25, W: 0, H: 0})
		}
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

func (g *Game) handlePlayerController(a int32) {

}

func (g *Game) drawPlayer(current *Snake) {
	if current.lost {
		return
	}
	for _, part := range current.parts {
		g.drawRect(part, current.color)
	}
}

func (g *Game) forward(current *Snake, index int) {
	if current.lost {
		return
	}

	head := current.parts[len(current.parts)-1]
	newHead := Point{x: head.x + current.direction.x,
		y: head.y + current.direction.y,
	}
	if config.CycleBorder {
		if newHead.x < 0 {
			newHead.x = config.Width - config.FieldSize
		}
		if newHead.y < 0 {
			newHead.y = config.Height - config.FieldSize
		}
		if newHead.x > config.Width {
			newHead.x = newHead.x - config.Width
		}
		if newHead.y > config.Height {
			newHead.y = newHead.y - config.Height
		}
	}
	current.parts = append(current.parts, newHead)

	foodHit := false
	for idx, f := range food {
		if newHead.x == f.x && newHead.y == f.y {
			foodHit = true
			food[idx] = g.randomPoint()
			current.score += 10
		}
	}
	if !foodHit {
		current.parts = current.parts[1:]
	}

	if newHead.x < 0 || newHead.y < 0 || newHead.x > config.Width || newHead.y > config.Height {
		current.lost = true
		//g.drawMessage(fmt.Sprintf("Player %s lost\n", current.name))
		//g.gameRunning = false
	}

	for i := 0; i < len(g.obstacles); i++ {
		if newHead == g.obstacles[i] {
			current.lost = true
			//g.drawMessage(fmt.Sprintf("Player %s lost\n", current.name))
			//g.gameRunning = false
		}
	}

	//collision with other player
	for idx, other := range players {
		for pidx, otherPart := range other.parts {
			if idx != index || pidx != len(current.parts)-1 {
				if otherPart == newHead {
					current.lost = false
					//g.drawMessage(fmt.Sprintf("Player %s lost\n", current.name))
					//g.gameRunning = false
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

func (g *Game) decideWinner() {
	singlePlayer := len(config.Players) == 1
	lastSurvived := -1
	activePlayers := 0
	for i := 0; i < len(players); i++ {
		if !players[i].lost {
			activePlayers++
			lastSurvived = i
		}
	}

	if singlePlayer && activePlayers == 0 {
		g.drawMessage(fmt.Sprintf("%s looses", players[0].name))
		g.gameRunning = false
	}
	if !singlePlayer && activePlayers <= 1 {
		if lastSurvived >= 0 {
			g.drawMessage(fmt.Sprintf("%s looses", players[lastSurvived].name))
		} else {
			//TODO decide on score
			g.drawMessage(fmt.Sprintf("%s looses", players[lastSurvived].name))
		}
		g.gameRunning = false
	}
}
