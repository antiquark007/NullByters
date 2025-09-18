package pages

import (
	"crypto/ed25519"
	"crypto/sha256"
	"data_wiper/internal/drivers"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	confirmClearActive bool = false
	clearTargetName    string
	clearConfirmText   string
	clearTextActive    bool    = false
	clearAnimationTime float32 = 0
)

const requiredClearText = "CLEAR"

func ShowConfirmClear(itemName string) {
	confirmClearActive = true
	clearTargetName = itemName
	clearConfirmText = ""
	clearTextActive = false
	clearAnimationTime = 0
}

func HideConfirmClear() {
	confirmClearActive = false
	clearTargetName = ""
	clearConfirmText = ""
	clearTextActive = false
	clearAnimationTime = 0
}

func IsConfirmClearActive() bool {
	return confirmClearActive
}

func DrawConfirmClear() {
	if !confirmClearActive {
		return
	}

	clearAnimationTime += rl.GetFrameTime()

	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	overlayAlpha := uint8(min(180, int(clearAnimationTime*300)))
	rl.DrawRectangle(0, 0, int32(screenWidth), int32(screenHeight),
		rl.NewColor(0, 0, 0, overlayAlpha))

	modalWidth := float32(500)
	modalHeight := float32(360)
	modalX := (screenWidth - modalWidth) / 2
	modalY := (screenHeight - modalHeight) / 2

	scale := float32(math.Min(1.0, float64(clearAnimationTime*4)))
	actualWidth := modalWidth * scale
	actualHeight := modalHeight * scale
	actualX := modalX + (modalWidth-actualWidth)/2
	actualY := modalY + (modalHeight-actualHeight)/2

	if scale < 0.1 {
		return
	}

	modalRect := rl.NewRectangle(actualX, actualY, actualWidth, actualHeight)

	rl.DrawRectangleGradientV(
		int32(actualX), int32(actualY), int32(actualWidth), int32(actualHeight),
		rl.NewColor(15, 25, 35, 250),
		rl.NewColor(5, 15, 25, 250),
	)

	rl.DrawRectangleRoundedLines(modalRect, 0.15, 8, rl.NewColor(255, 180, 0, 255))

	glowRect := rl.NewRectangle(actualX-2, actualY-2, actualWidth+4, actualHeight+4)
	rl.DrawRectangleRounded(glowRect, 0.15, 12, rl.NewColor(255, 180, 0, 30))

	if scale < 1.0 {
		return 
	}

	
	headerHeight := float32(60)
	headerRect := rl.NewRectangle(modalX, modalY, modalWidth, headerHeight)
	rl.DrawRectangleRounded(headerRect, 0.15, 8, rl.NewColor(25, 35, 45, 200))

	iconSize := float32(24)
	iconX := modalX + 20
	iconY := modalY + (headerHeight-iconSize)/2

	rl.DrawCircle(int32(iconX+iconSize/2), int32(iconY+iconSize/2), iconSize/2, rl.NewColor(0, 255, 180, 255))
	rl.DrawCircleLines(int32(iconX+iconSize/2), int32(iconY+iconSize/2), iconSize/2, rl.NewColor(50, 255, 200, 255))

	rl.DrawText("i", int32(iconX+iconSize/2-3), int32(iconY+4), 16, rl.NewColor(5, 15, 20, 255))

	
	titleText := "Confirm Clear Operation"
	rl.DrawText(titleText, int32(iconX+iconSize+15), int32(modalY+20), 20, rl.NewColor(0, 255, 180, 255))

	closeSize := float32(30)
	closeX := modalX + modalWidth - closeSize - 15
	closeY := modalY + 15
	closeRect := rl.NewRectangle(closeX, closeY, closeSize, closeSize)

	mouse := rl.GetMousePosition()
	closeHover := rl.CheckCollisionPointRec(mouse, closeRect)

	if closeHover {
		rl.DrawRectangleRounded(closeRect, 0.3, 6, rl.NewColor(0, 255, 180, 100))
	}

	rl.DrawText("×", int32(closeX+8), int32(closeY+2), 24, rl.NewColor(200, 200, 200, 255))

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && closeHover {
		HideConfirmClear()
		return
	}

	contentY := modalY + headerHeight + 20

	infoText := "This will clear the contents of the item but keep"
	rl.DrawText(infoText, int32(modalX+20), int32(contentY), 16, rl.NewColor(200, 200, 200, 255))

	infoText2 := "the file/folder structure intact:"
	rl.DrawText(infoText2, int32(modalX+20), int32(contentY+22), 16, rl.NewColor(200, 200, 200, 255))

	targetY := contentY + 55
	targetRect := rl.NewRectangle(modalX+20, targetY, modalWidth-40, 40)
	rl.DrawRectangleRounded(targetRect, 0.1, 6, rl.NewColor(20, 40, 30, 255))
	rl.DrawRectangleRoundedLines(targetRect, 0.1, 1, rl.NewColor(0, 255, 180, 255))

	
	displayName := clearTargetName
	maxChars := 55
	if len(displayName) > maxChars {
		displayName = displayName[:maxChars-3] + "..."
	}
	rl.DrawText(displayName, int32(modalX+30), int32(targetY+12), 16, rl.NewColor(100, 255, 200, 255))

	instructionY := targetY + 60
	instructionText := fmt.Sprintf("Please type %s to confirm:", requiredClearText)
	rl.DrawText(instructionText, int32(modalX+20), int32(instructionY), 16, rl.NewColor(200, 200, 200, 255))

	inputY := instructionY + 30
	inputRect := rl.NewRectangle(modalX+20, inputY, modalWidth-40, 40)

	inputHover := rl.CheckCollisionPointRec(mouse, inputRect)
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && inputHover {
		clearTextActive = true
	} else if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !inputHover {
		clearTextActive = false
	}

	var inputBg rl.Color
	var inputBorder rl.Color
	if clearTextActive {
		inputBg = rl.NewColor(25, 35, 45, 255)
		inputBorder = rl.NewColor(0, 255, 180, 255)
	} else if inputHover {
		inputBg = rl.NewColor(20, 30, 40, 255)
		inputBorder = rl.NewColor(100, 200, 150, 255)
	} else {
		inputBg = rl.NewColor(15, 25, 35, 255)
		inputBorder = rl.NewColor(60, 120, 90, 255)
	}

	rl.DrawRectangleRounded(inputRect, 0.1, 6, inputBg)
	rl.DrawRectangleRoundedLines(inputRect, 0.1, 1, inputBorder)

	if clearTextActive {
		if rl.IsKeyPressed(rl.KeyBackspace) && len(clearConfirmText) > 0 {
			clearConfirmText = clearConfirmText[:len(clearConfirmText)-1]
		}

		key := rl.GetCharPressed()
		for key > 0 {
			if key >= 32 && key <= 125 && len(clearConfirmText) < 20 {
				clearConfirmText += strings.ToUpper(string(rune(key)))
			}
			key = rl.GetCharPressed()
		}
	}

	textColor := rl.NewColor(255, 255, 255, 255)
	if clearConfirmText == requiredClearText {
		textColor = rl.NewColor(100, 255, 100, 255)
	} else if len(clearConfirmText) > 0 {
		textColor = rl.NewColor(255, 180, 100, 255)
	}

	inputTextX := modalX + 30
	inputTextY := inputY + 12
	rl.DrawText(clearConfirmText, int32(inputTextX), int32(inputTextY), 16, textColor)

	if clearTextActive && int(rl.GetTime()*2)%2 == 0 {
		cursorX := inputTextX + float32(rl.MeasureText(clearConfirmText, 16))
		rl.DrawText("|", int32(cursorX), int32(inputTextY), 16, rl.NewColor(0, 255, 180, 255))
	}

	buttonY := inputY + 60
	buttonHeight := float32(35)

	cancelWidth := float32(80)
	cancelRect := rl.NewRectangle(modalX+modalWidth-cancelWidth-120, buttonY, cancelWidth, buttonHeight)
	cancelHover := rl.CheckCollisionPointRec(mouse, cancelRect)

	cancelBg := rl.NewColor(60, 60, 60, 255)
	cancelBorder := rl.NewColor(120, 120, 120, 255)
	if cancelHover {
		cancelBg = rl.NewColor(80, 80, 80, 255)
		cancelBorder = rl.NewColor(160, 160, 160, 255)
	}

	rl.DrawRectangleRounded(cancelRect, 0.2, 6, cancelBg)
	rl.DrawRectangleRoundedLines(cancelRect, 0.2, 1, cancelBorder)
	rl.DrawText("Cancel", int32(cancelRect.X+20), int32(cancelRect.Y+9), 16, rl.NewColor(255, 255, 255, 255))

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && cancelHover {
		HideConfirmClear()
		return
	}

	clearWidth := float32(100)
	clearRect := rl.NewRectangle(modalX+modalWidth-clearWidth-20, buttonY, clearWidth, buttonHeight)
	clearHover := rl.CheckCollisionPointRec(mouse, clearRect)
	canClear := clearConfirmText == requiredClearText

	var clearBg rl.Color
	var clearBorder rl.Color
	var clearTextColor rl.Color

	if !canClear {
		clearBg = rl.NewColor(30, 60, 40, 255)
		clearBorder = rl.NewColor(50, 100, 60, 255)
		clearTextColor = rl.NewColor(100, 150, 120, 255)
	} else if clearHover {
		clearBg = rl.NewColor(50, 255, 200, 255)
		clearBorder = rl.NewColor(100, 255, 220, 255)
		clearTextColor = rl.NewColor(5, 15, 20, 255)
		glowClear := rl.NewRectangle(clearRect.X-1, clearRect.Y-1, clearRect.Width+2, clearRect.Height+2)
		rl.DrawRectangleRounded(glowClear, 0.2, 8, rl.NewColor(50, 255, 200, 80))
	} else {
		clearBg = rl.NewColor(0, 255, 180, 255)
		clearBorder = rl.NewColor(50, 255, 200, 255)
		clearTextColor = rl.NewColor(5, 15, 20, 255)
	}

	rl.DrawRectangleRounded(clearRect, 0.2, 6, clearBg)
	rl.DrawRectangleRoundedLines(clearRect, 0.2, 1, clearBorder)
	rl.DrawText("Clear Item", int32(clearRect.X+15), int32(clearRect.Y+9), 16, clearTextColor)

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && clearHover && canClear {
		start := time.Now()
		err := drivers.ClearItem(clearTargetName)
		status := "success"
		if err != nil {
			status = "failure"
			fmt.Printf("Clear failed: %v\n", err)
		}
		finished := time.Now()
		duration := int(finished.Sub(start).Seconds())

		var log WipeLog
		log.Wipe.Method = "overwrite"
		log.Wipe.NistLevel = "clear"
		log.Wipe.Status = status
		log.Wipe.StartedAt = start.UTC().Format(time.RFC3339)
		log.Wipe.FinishedAt = finished.UTC().Format(time.RFC3339)
		log.Wipe.DurationSec = duration
		log.System.ToolVersion = "v1.0"
		log.System.HostOS = "Ubuntu 22.04"
		log.System.ExecutedBy = os.Getenv("USER")

		isDevice := strings.HasPrefix(clearTargetName, "/dev/")
		if isDevice {
			devInfo, err := getDeviceInfo(clearTargetName)
			if err == nil {
				log.Device.Name = devInfo["name"].(string)
				log.Device.Serial = devInfo["serial"].(string)
				log.Device.SizeGB = devInfo["size_gb"].(int)
				log.Device.Type = devInfo["type"].(string)
			}
		} else {
			fi, err := os.Stat(clearTargetName)
			if err == nil {
				log.Device.Name = clearTargetName
				log.Device.Serial = ""
				log.Device.SizeGB = int(fi.Size() / 1000000000)
				log.Device.Type = "file"
			}
		}

		
		temp := struct {
			Device struct {
				Name   string `json:"name"`
				Serial string `json:"serial"`
				SizeGB int    `json:"size_gb"`
				Type   string `json:"type"`
			} `json:"device"`
			Wipe struct {
				Method      string `json:"method"`
				NistLevel   string `json:"nist_level"`
				Status      string `json:"status"`
				StartedAt   string `json:"started_at"`
				FinishedAt  string `json:"finished_at"`
				DurationSec int    `json:"duration_sec"`
			} `json:"wipe"`
			System struct {
				ToolVersion string `json:"tool_version"`
				HostOS      string `json:"host_os"`
				ExecutedBy  string `json:"executed_by"`
			} `json:"system"`
		}{
			Device: log.Device,
			Wipe:   log.Wipe,
			System: log.System,
		}
		jsonBytes, _ := json.Marshal(temp)
		sig := ed25519.Sign(privateKey, jsonBytes)
		log.Signature.Algorithm = "Ed25519"
		log.Signature.Sig = base64.StdEncoding.EncodeToString(sig)
		hash := sha256.Sum256(publicKey)
		log.Signature.PublicKeyFingerprint = fmt.Sprintf("%x", hash[:])

		ShowCertificate(log)
		HideConfirmClear()
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		HideConfirmClear()
	}
	if rl.IsKeyPressed(rl.KeyEnter) && canClear {
		start := time.Now()
		err := drivers.ClearItem(clearTargetName)
		status := "success"
		if err != nil {
			status = "failure"
			fmt.Printf("Clear failed: %v\n", err)
		}
		finished := time.Now()
		duration := int(finished.Sub(start).Seconds())

		var log WipeLog
		log.Wipe.Method = "overwrite"
		log.Wipe.NistLevel = "clear"
		log.Wipe.Status = status
		log.Wipe.StartedAt = start.UTC().Format(time.RFC3339)
		log.Wipe.FinishedAt = finished.UTC().Format(time.RFC3339)
		log.Wipe.DurationSec = duration
		log.System.ToolVersion = "v1.0"
		log.System.HostOS = "Ubuntu 22.04"
		log.System.ExecutedBy = os.Getenv("USER")

		isDevice := strings.HasPrefix(clearTargetName, "/dev/")
		if isDevice {
			devInfo, err := getDeviceInfo(clearTargetName)
			if err == nil {
				log.Device.Name = devInfo["name"].(string)
				log.Device.Serial = devInfo["serial"].(string)
				log.Device.SizeGB = devInfo["size_gb"].(int)
				log.Device.Type = devInfo["type"].(string)
			}
		} else {
			fi, err := os.Stat(clearTargetName)
			if err == nil {
				log.Device.Name = clearTargetName
				log.Device.Serial = ""
				log.Device.SizeGB = int(fi.Size() / 1000000000)
				log.Device.Type = "file"
			}
		}

		
		temp := struct {
			Device struct {
				Name   string `json:"name"`
				Serial string `json:"serial"`
				SizeGB int    `json:"size_gb"`
				Type   string `json:"type"`
			} `json:"device"`
			Wipe struct {
				Method      string `json:"method"`
				NistLevel   string `json:"nist_level"`
				Status      string `json:"status"`
				StartedAt   string `json:"started_at"`
				FinishedAt  string `json:"finished_at"`
				DurationSec int    `json:"duration_sec"`
			} `json:"wipe"`
			System struct {
				ToolVersion string `json:"tool_version"`
				HostOS      string `json:"host_os"`
				ExecutedBy  string `json:"executed_by"`
			} `json:"system"`
		}{
			Device: log.Device,
			Wipe:   log.Wipe,
			System: log.System,
		}
		jsonBytes, _ := json.Marshal(temp)
		sig := ed25519.Sign(privateKey, jsonBytes)
		log.Signature.Algorithm = "Ed25519"
		log.Signature.Sig = base64.StdEncoding.EncodeToString(sig)
		hash := sha256.Sum256(publicKey)
		log.Signature.PublicKeyFingerprint = fmt.Sprintf("%x", hash[:])

		ShowCertificate(log)
		HideConfirmClear()
	}
}
