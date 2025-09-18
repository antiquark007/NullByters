package ui

import (
	// "data_wiper/internal/pages"

	"data_wiper/internal/pages"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	PageLogin = iota
	PageDashboard
	PageSettings
)

var currentPage = PageLogin

func RunUI() {
	rl.InitWindow(800, 600, "Secure Wipe Tool")
	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		switch currentPage {
		case PageLogin:
			if pages.DrawLogin() { // return true jab login successful ho
				currentPage = PageDashboard
			}
		case PageDashboard:
			pages.DrawDashboard()

		}

		rl.EndDrawing()
	}
}
