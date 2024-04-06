// Fronton 2D game with multiple balls (3 or more), similar to famous Pong.
// Divertimento in go with ebiten. Thanks to GO team and Hajimehoshi.
// Marc Riart, 202404.

package main

import (
	"fmt"
	"image/color"
	"math/rand/v2"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	windowTitle  = "Fronton"
	screenWidth  = 500
	screenHeight = 700
	racketWidth  = screenWidth / 4
	racketHeight = 5 // +5 pixels of distance between racket and floor, fixed
	racketSpeed  = 8
	winnerScore  = 21
)

const (
	audioStart = iota
	audioHit
	audioMiss
	audioOver
)

type Game struct {
	State int // Defines the game state: 0 not initiated, 1 started, 2 over
	Score struct {
		Player int
		CPU    int
	}
	Balls  []Ball
	Racket struct {
		X     int // The x point of the vertex top left
		Y     int // The y point of the vertex. It is a fixed value, but useful for geometry
		Speed int
	}
	audioContext     *audio.Context
	audioPlayerStart *audio.Player
	audioPlayerHit   *audio.Player
	audioPlayerMiss  *audio.Player
	audioPlayerOver  *audio.Player
}

type Ball struct {
	Radius int
	Color  color.RGBA
	X      int
	Y      int
	SpeedX int
	SpeedY int
}

func (g *Game) Initialize() {
	g.State = 0
	g.Score.Player = 0
	g.Score.CPU = 0

	g.Balls = []Ball{}

	g.Racket.X = screenWidth/2 - racketWidth/2
	g.Racket.Y = screenHeight - 1 - 5 - racketHeight
	g.Racket.Speed = racketSpeed

	if g.audioContext == nil {
		g.audioContext = audio.NewContext(48000)

		// Audio start
		fStart, err := os.Open("res/game-start-6104.mp3")
		if err != nil {
			panic(err)
		}
		dStart, _ := mp3.DecodeWithoutResampling(fStart)
		g.audioPlayerStart, _ = g.audioContext.NewPlayer(dStart)

		// Audio hit the ball
		fHit, err := os.Open("res/one_beep-99630.mp3")
		if err != nil {
			panic(err)
		}
		dHit, _ := mp3.DecodeWithoutResampling(fHit)
		g.audioPlayerHit, _ = g.audioContext.NewPlayer(dHit)

		// Audio misses the ball
		fMiss, err := os.Open("res/coin-collect-retro-8-bit-sound-effect-145251.mp3")
		if err != nil {
			panic(err)
		}
		dMiss, _ := mp3.DecodeWithoutResampling(fMiss)
		g.audioPlayerMiss, _ = g.audioContext.NewPlayer(dMiss)

		// Audio game over
		fOver, err := os.Open("res/cute-level-up-3-189853.mp3")
		if err != nil {
			panic(err)
		}
		dOver, _ := mp3.DecodeWithoutResampling(fOver)
		g.audioPlayerOver, _ = g.audioContext.NewPlayer(dOver)
	}
}

func (g *Game) AddBall(rad int, R, G, B, A uint8, speedx, speedy int) {
	g.Balls = append(g.Balls, Ball{})
	last := len(g.Balls) - 1
	g.Balls[last].Radius = rad
	g.Balls[last].Color = color.RGBA{R, G, B, A}
	g.Balls[last].X = rand.IntN(screenWidth)
	g.Balls[last].Y = rad
	g.Balls[last].SpeedX = speedx
	g.Balls[last].SpeedY = speedy
}

func (b *Ball) ResetBall() {
	b.X, b.Y = rand.IntN(screenWidth), b.Radius
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {
	// Logic for start the game over. Space to start
	if g.State == 0 {
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			os.Exit(0)
		}

		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.State = 1
			go g.PlaySound(audioStart)
		}

		return nil
	}

	// Logic for exit. Escape to finish
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}

	// Logic for re-start once game over. Space to re-start
	if g.State == 2 {
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			os.Exit(0)
		}

		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.Initialize()
			g.AddBall(5, 0, 255, 0, 0, 5, 5)
		}

		return nil
	}

	// Logic for the match, g.State = 1

	// Move the racket
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.Racket.X -= g.Racket.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.Racket.X += g.Racket.Speed
	}

	// Move the balls. Loop over each ball
	for i, _ := range g.Balls {
		// Move the ball b
		g.Balls[i].X += g.Balls[i].SpeedX
		g.Balls[i].Y += g.Balls[i].SpeedY

		// Ball out of the field lines, but not hitting racket or missing yet
		if (g.Balls[i].Y + 5) < g.Racket.Y {
			// Ball hits the walls
			if (g.Balls[i].X-g.Balls[i].Radius <= 0) || (g.Balls[i].X+g.Balls[i].Radius) >= (screenWidth-1) {
				g.Balls[i].SpeedX = -g.Balls[i].SpeedX
			}

			// Ball hits the ceil
			if (g.Balls[i].Y - g.Balls[i].Radius) <= 0 {
				g.Balls[i].SpeedY = -g.Balls[i].SpeedY
			}

			continue
		}

		// Ball below the field line, hits or misses the racket
		// Hits the racket, else misses
		if g.Balls[i].X >= g.Racket.X && g.Balls[i].X <= (g.Racket.X+racketWidth) {
			//g.Balls[i].SpeedX = accelerate(g.Balls[i].SpeedX)
			//g.Balls[i].SpeedY = accelerateNRevers(g.Balls[i].SpeedY)
			g.Balls[i].SpeedY = -g.Balls[i].SpeedY
			g.Score.Player++
			go g.PlaySound(audioHit)
		} else {
			g.Balls[i].ResetBall()
			g.Score.CPU++
			go g.PlaySound(audioMiss)
		}
	}

	// Review scores and act accordingly:
	// -If reached winnerScore, game is over
	// -Add a new ball every 3 player points
	if g.Score.Player == winnerScore || g.Score.CPU == winnerScore {
		go g.PlaySound(audioOver)
		g.State = 2
	} else if g.Score.Player == len(g.Balls)*3 {
		// Add a total random ball
		speed := randBetween(3, 6)
		g.AddBall(randBetween(4, 10), uint8(randBetween(0, 255)), uint8(randBetween(0, 255)), uint8(randBetween(0, 255)), 0, speed, speed)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw initial screen, pre-start
	if g.State == 0 {
		vector.DrawFilledCircle(screen, screenWidth/2, screenHeight/2, 5, color.RGBA{R: 0, G: 255, B: 0, A: 0}, true)
		vector.DrawFilledRect(screen, float32(g.Racket.X), screenHeight-racketHeight-5, racketWidth, racketHeight, color.White, false)
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Press SPACE to star, ESC at any time to finish.\nMove racket with horizontal arrows.\nFirst to score %d wins. Enjoy!", winnerScore))
		return
	}

	// Draw game over
	if g.State == 2 {
		if g.Score.Player == winnerScore {
			ebitenutil.DebugPrint(screen, "Game over. You won!")
		} else {
			ebitenutil.DebugPrint(screen, "Game over. I won!")
		}
		ebitenutil.DebugPrint(screen, "\n\nPress SPACE to star again, ESC to exit.")
		return
	}

	// Draw the match

	// Draw balls
	for _, b := range g.Balls {
		vector.DrawFilledCircle(screen, float32(b.X), float32(b.Y), float32(b.Radius), b.Color, true)
	}
	// Draw racket
	vector.DrawFilledRect(screen, float32(g.Racket.X), screenHeight-racketHeight-5, racketWidth, racketHeight, color.White, false)

	// Print score
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Score: %d/%d", g.Score.Player, g.Score.CPU))
}

// Return a random integer between n and m, both included
func randBetween(n, m int) int {
	if !(m > n) {
		return 0
	}
	return rand.IntN(m+1-n) + n
}

// Play a sound for each situation (sit)
func (g *Game) PlaySound(sit int) {
	switch sit {
	case audioStart:
		g.audioPlayerStart.Rewind()
		g.audioPlayerStart.Play()
	case audioHit:
		g.audioPlayerHit.Rewind()
		g.audioPlayerHit.Play()
	case audioMiss:
		g.audioPlayerMiss.Rewind()
		g.audioPlayerMiss.Play()
	case audioOver:
		g.audioPlayerOver.Rewind()
		g.audioPlayerOver.Play()
	}
}

func main() {
	g := Game{}
	g.Initialize()
	g.AddBall(5, 0, 255, 0, 0, 5, 5)

	ebiten.SetWindowTitle(windowTitle)
	ebiten.SetWindowSize(screenWidth, screenHeight)

	err := ebiten.RunGame(&g)
	if err != nil {
		panic(err)
	}
}
