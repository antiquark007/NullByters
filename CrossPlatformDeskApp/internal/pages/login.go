package pages

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var username, password string
var loginSuccess bool
var activeField int
var usernameRect, passwordRect, loginButtonRect rl.Rectangle
var showPassword bool = false
var loginAttempts int = 0
var errorMessage string
var errorTimer float32 = 0

func DrawLogin() bool {
	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	formWidth := float32(400)
	formHeight := float32(350)
	startX := (screenWidth - formWidth) / 2
	startY := (screenHeight-formHeight)/2 - 50

	fieldWidth := float32(320)
	fieldHeight := float32(45)
	fieldX := startX + (formWidth-fieldWidth)/2

	usernameRect = rl.NewRectangle(fieldX, startY+80, fieldWidth, fieldHeight)
	passwordRect = rl.NewRectangle(fieldX, startY+160, fieldWidth-50, fieldHeight)
	showPasswordRect := rl.NewRectangle(fieldX+fieldWidth-45, startY+160, 40, fieldHeight)
	loginButtonRect = rl.NewRectangle(fieldX+(fieldWidth-120)/2, startY+240, 120, 45)

	rl.DrawRectangleGradientV(
		0, 0, int32(screenWidth), int32(screenHeight),
		rl.NewColor(5, 15, 20, 255),
		rl.NewColor(0, 200, 120, 255),
	)

	for i := 0; i < 80; i++ {
		x := float32((i*47 + int(rl.GetTime()*20)) % int(screenWidth))
		y := float32((i*89 + int(rl.GetTime()*25)) % int(screenHeight))
		alpha := uint8(100 + int(50*(1+float32(i%3))))
		rl.DrawCircle(int32(x), int32(y), 1.5, rl.NewColor(0, 255, 180, alpha))
	}

	containerRect := rl.NewRectangle(startX-30, startY-30, formWidth+60, formHeight+80)
	rl.DrawRectangleRounded(containerRect, 0.15, 12, rl.NewColor(10, 40, 25, 220))
	rl.DrawRectangleRoundedLines(containerRect, 0.15, 2, rl.NewColor(0, 255, 180, 150))

	titleText := "Null Byters - Login"
	titleSize := int32(28)
	titleWidth := rl.MeasureText(titleText, titleSize)
	titleX := int32(screenWidth/2 - float32(titleWidth)/2)
	rl.DrawText(titleText, titleX, int32(startY-10), titleSize, rl.NewColor(0, 255, 180, 255))

	subtitleText := "Secure Data Management System"
	subtitleSize := int32(14)
	subtitleWidth := rl.MeasureText(subtitleText, subtitleSize)
	subtitleX := int32(screenWidth/2 - float32(subtitleWidth)/2)
	rl.DrawText(subtitleText, subtitleX, int32(startY+25), subtitleSize, rl.NewColor(100, 200, 150, 200))

	drawInputField(usernameRect, "Username:", username, activeField == 1, false)

	drawInputField(passwordRect, "Password:", password, activeField == 2, !showPassword)

	toggleColor := rl.NewColor(0, 255, 180, 200)
	if showPassword {
		toggleColor = rl.NewColor(0, 255, 180, 255)
	}

	rl.DrawRectangleRounded(showPasswordRect, 0.2, 6, rl.NewColor(15, 60, 40, 180))
	rl.DrawRectangleRoundedLines(showPasswordRect, 0.2, 1, toggleColor)
	eyeText := "👁"
	if showPassword {
		eyeText = "🙈"
	}
	rl.DrawText(eyeText, int32(showPasswordRect.X+12), int32(showPasswordRect.Y+12), 20, toggleColor)

	mouse := rl.GetMousePosition()
	buttonHover := rl.CheckCollisionPointRec(mouse, loginButtonRect)
	canLogin := len(username) > 0 && len(password) > 0

	var buttonColor rl.Color
	var textColor rl.Color

	if !canLogin {
		buttonColor = rl.NewColor(60, 60, 60, 200)
		textColor = rl.NewColor(120, 120, 120, 255)
	} else if buttonHover {
		buttonColor = rl.NewColor(0, 255, 180, 255)
		textColor = rl.NewColor(5, 15, 20, 255)

		glowRect := rl.NewRectangle(loginButtonRect.X-2, loginButtonRect.Y-2, loginButtonRect.Width+4, loginButtonRect.Height+4)
		rl.DrawRectangleRounded(glowRect, 0.2, 8, rl.NewColor(0, 255, 180, 100))
	} else {
		buttonColor = rl.NewColor(0, 200, 140, 255)
		textColor = rl.NewColor(255, 255, 255, 255)
	}

	rl.DrawRectangleRounded(loginButtonRect, 0.2, 8, buttonColor)
	if canLogin {
		rl.DrawRectangleRoundedLines(loginButtonRect, 0.2, 2, rl.NewColor(50, 255, 200, 255))
	}

	loginText := "LOGIN"
	loginTextWidth := rl.MeasureText(loginText, 18)
	loginTextX := loginButtonRect.X + (loginButtonRect.Width-float32(loginTextWidth))/2
	loginTextY := loginButtonRect.Y + (loginButtonRect.Height-18)/2
	rl.DrawText(loginText, int32(loginTextX), int32(loginTextY), 18, textColor)

	if errorTimer > 0 {
		errorTimer -= rl.GetFrameTime()
		alpha := uint8(255 * (errorTimer / 3.0))
		if alpha > 0 {
			errorWidth := rl.MeasureText(errorMessage, 16)
			errorX := int32(screenWidth/2 - float32(errorWidth)/2)
			errorY := int32(startY + 300)

			errorRect := rl.NewRectangle(float32(errorX)-15, float32(errorY)-8, float32(errorWidth)+30, 32)
			rl.DrawRectangleRounded(errorRect, 0.2, 6, rl.NewColor(200, 50, 50, alpha/3))
			rl.DrawRectangleRoundedLines(errorRect, 0.2, 1, rl.NewColor(255, 100, 100, alpha))

			rl.DrawText(errorMessage, errorX, errorY, 16, rl.NewColor(255, 150, 150, alpha))
		}
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if rl.CheckCollisionPointRec(mouse, usernameRect) {
			activeField = 1
		} else if rl.CheckCollisionPointRec(mouse, passwordRect) {
			activeField = 2
		} else if rl.CheckCollisionPointRec(mouse, showPasswordRect) {
			showPassword = !showPassword
		} else if rl.CheckCollisionPointRec(mouse, loginButtonRect) && canLogin {
			return attemptLogin()
		} else {
			activeField = 0
		}
	}

	if activeField > 0 {

		key := rl.GetCharPressed()
		for key > 0 {
			if key >= 32 && key <= 125 {
				if activeField == 1 && len(username) < 25 {
					username += string(rune(key))
				} else if activeField == 2 && len(password) < 30 {
					password += string(rune(key))
				}
			}
			key = rl.GetCharPressed()
		}

		if rl.IsKeyPressed(rl.KeyBackspace) {
			if activeField == 1 && len(username) > 0 {
				username = username[:len(username)-1]
			} else if activeField == 2 && len(password) > 0 {
				password = password[:len(password)-1]
			}
		}

		if rl.IsKeyPressed(rl.KeyTab) {
			if activeField == 1 {
				activeField = 2
			} else {
				activeField = 1
			}
		}
	}

	if rl.IsKeyPressed(rl.KeyEnter) && canLogin {
		return attemptLogin()
	}

	instructionY := int32(screenHeight - 60)
	rl.DrawText("Use Tab to switch fields • Enter to login", int32(screenWidth/2-150), instructionY, 14, rl.NewColor(100, 200, 150, 180))

	return false
}

func drawInputField(rect rl.Rectangle, label string, text string, isActive bool, maskText bool) {

	rl.DrawText(label, int32(rect.X), int32(rect.Y-25), 16, rl.NewColor(0, 255, 180, 255))

	var bgColor rl.Color
	var borderColor rl.Color

	if isActive {
		bgColor = rl.NewColor(25, 80, 50, 220)
		borderColor = rl.NewColor(50, 255, 200, 255)
	} else {
		bgColor = rl.NewColor(15, 60, 40, 200)
		borderColor = rl.NewColor(0, 255, 180, 150)
	}

	rl.DrawRectangleRounded(rect, 0.2, 8, bgColor)
	borderWidth := 1
	if isActive {
		borderWidth = 2
	}
	rl.DrawRectangleRoundedLines(rect, 0.2, int32(borderWidth), borderColor)

	displayText := text
	if maskText && len(text) > 0 {
		displayText = ""
		for i := 0; i < len(text); i++ {
			displayText += "•"
		}
	}

	maxWidth := int(rect.Width - 20)
	textWidth := rl.MeasureText(displayText, 18)
	if textWidth > int32(maxWidth) {

		for len(displayText) > 0 && rl.MeasureText(displayText, 18) > int32(maxWidth) {
			displayText = displayText[1:]
		}
	}

	textColor := rl.NewColor(255, 255, 255, 255)
	if len(displayText) == 0 && !isActive {

		placeholderText := "Enter " + strings.ToLower(label[:len(label)-1]) + "..."
		rl.DrawText(placeholderText, int32(rect.X+12), int32(rect.Y+13), 16, rl.NewColor(120, 180, 140, 180))
	} else {
		rl.DrawText(displayText, int32(rect.X+12), int32(rect.Y+13), 18, textColor)
	}

	if isActive && int(rl.GetTime()*2)%2 == 0 {
		cursorX := rect.X + 12 + float32(rl.MeasureText(displayText, 18))
		rl.DrawText("|", int32(cursorX), int32(rect.Y+13), 18, rl.NewColor(0, 255, 180, 255))
	}
}

func attemptLogin() bool {
	// Simple validation (you can enhance this with real authentication)
	if len(username) == 0 {
		showError("Username cannot be empty")
		return false
	}

	if len(password) < 4 {
		showError("Password must be at least 4 characters")
		return false
	}

	if len(username) > 0 && len(password) >= 4 {
		loginSuccess = true
		return true
	}

	loginAttempts++
	if loginAttempts >= 3 {
		showError("Too many failed attempts. Please wait...")
	} else {
		showError("Invalid credentials. Please try again.")
	}

	return false
}

func showError(message string) {
	errorMessage = message
	errorTimer = 3.0
}
