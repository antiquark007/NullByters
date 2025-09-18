package main

import (r1 "github.com/gen2brain/raylib-go/raylib"
        "data_wiper/internal/ui"
)

func main() {
	r1.InitWindow(880, 600, "Secure Byters")
	defer r1.CloseWindow()
	for !r1.WindowShouldClose() {
		r1.BeginDrawing()
		r1.ClearBackground(r1.RayWhite)
		ui.RunUI()
		r1.EndDrawing()
		 
		
	}
}
