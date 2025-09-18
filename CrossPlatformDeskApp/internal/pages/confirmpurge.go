package pages

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
    "data_wiper/internal/drivers"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	confirmPurgeActive bool = false
	purgeTargetName    string
	purgeConfirmText   string
	purgeTextActive    bool    = false
	purgeAnimationTime float32 = 0
)

const requiredPurgeText = "DELETE"

func ShowConfirmPurge(itemName string) {
	confirmPurgeActive = true
	purgeTargetName = itemName
	purgeConfirmText = ""
	purgeTextActive = false
	purgeAnimationTime = 0
}

func HideConfirmPurge() {
	confirmPurgeActive = false
	purgeTargetName = ""
	purgeConfirmText = ""
	purgeTextActive = false
	purgeAnimationTime = 0
}

func IsConfirmPurgeActive() bool {
	return confirmPurgeActive
}

func DrawConfirmPurge() {
	if !confirmPurgeActive {
		return
	}

	purgeAnimationTime += rl.GetFrameTime()

	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	overlayAlpha := uint8(min(180, int(purgeAnimationTime*300)))
	rl.DrawRectangle(0, 0, int32(screenWidth), int32(screenHeight),
		rl.NewColor(0, 0, 0, overlayAlpha))

	modalWidth := float32(520)
	modalHeight := float32(380)
	modalX := (screenWidth - modalWidth) / 2
	modalY := (screenHeight - modalHeight) / 2

	scale := float32(math.Min(1.0, float64(purgeAnimationTime*4)))
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

	rl.DrawRectangleRoundedLines(modalRect, 0.15, 8, rl.NewColor(220, 50, 50, 255))

	glowRect := rl.NewRectangle(actualX-2, actualY-2, actualWidth+4, actualHeight+4)
	rl.DrawRectangleRounded(glowRect, 0.15, 12, rl.NewColor(220, 50, 50, 30))

	if scale < 1.0 {
		return
	}

	
	headerHeight := float32(60)
	headerRect := rl.NewRectangle(modalX, modalY, modalWidth, headerHeight)
	rl.DrawRectangleRounded(headerRect, 0.15, 8, rl.NewColor(25, 35, 45, 200))

	iconSize := float32(24)
	iconX := modalX + 20
	iconY := modalY + (headerHeight-iconSize)/2

	
	triY := iconY + iconSize
	triTop := rl.NewVector2(iconX+iconSize/2, iconY)
	triLeft := rl.NewVector2(iconX, triY)
	triRight := rl.NewVector2(iconX+iconSize, triY)
	rl.DrawTriangle(triTop, triLeft, triRight, rl.NewColor(255, 180, 0, 255))

	rl.DrawText("!", int32(iconX+iconSize/2-3), int32(iconY+4), 16, rl.NewColor(25, 25, 25, 255))

	// Title
	titleText := "Confirm Purge Operation"
	rl.DrawText(titleText, int32(iconX+iconSize+15), int32(modalY+20), 20, rl.NewColor(255, 100, 100, 255))

	closeSize := float32(30)
	closeX := modalX + modalWidth - closeSize - 15
	closeY := modalY + 15
	closeRect := rl.NewRectangle(closeX, closeY, closeSize, closeSize)

	mouse := rl.GetMousePosition()
	closeHover := rl.CheckCollisionPointRec(mouse, closeRect)

	if closeHover {
		rl.DrawRectangleRounded(closeRect, 0.3, 6, rl.NewColor(255, 100, 100, 100))
	}

	rl.DrawText("×", int32(closeX+8), int32(closeY+2), 24, rl.NewColor(200, 200, 200, 255))

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && closeHover {
		HideConfirmPurge()
		return
	}

	contentY := modalY + headerHeight + 20

	warningText := "This action cannot be undone. This will permanently"
	rl.DrawText(warningText, int32(modalX+20), int32(contentY), 16, rl.NewColor(200, 200, 200, 255))

	warningText2 := "purge and destroy all data in:"
	rl.DrawText(warningText2, int32(modalX+20), int32(contentY+22), 16, rl.NewColor(200, 200, 200, 255))

	targetY := contentY + 55
	targetRect := rl.NewRectangle(modalX+20, targetY, modalWidth-40, 40)
	rl.DrawRectangleRounded(targetRect, 0.1, 6, rl.NewColor(40, 20, 20, 255))
	rl.DrawRectangleRoundedLines(targetRect, 0.1, 1, rl.NewColor(220, 50, 50, 255))

	displayName := purgeTargetName
	maxChars := 60
	if len(displayName) > maxChars {
		displayName = displayName[:maxChars-3] + "..."
	}
	rl.DrawText(displayName, int32(modalX+30), int32(targetY+12), 16, rl.NewColor(255, 150, 150, 255))

	instructionY := targetY + 60
	instructionText := fmt.Sprintf("Please type %s to confirm:", requiredPurgeText)
	rl.DrawText(instructionText, int32(modalX+20), int32(instructionY), 16, rl.NewColor(200, 200, 200, 255))

	inputY := instructionY + 30
	inputRect := rl.NewRectangle(modalX+20, inputY, modalWidth-40, 40)

	inputHover := rl.CheckCollisionPointRec(mouse, inputRect)
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && inputHover {
		purgeTextActive = true
	} else if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !inputHover {
		purgeTextActive = false
	}

	var inputBg rl.Color
	var inputBorder rl.Color
	if purgeTextActive {
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

	if purgeTextActive {
		if rl.IsKeyPressed(rl.KeyBackspace) && len(purgeConfirmText) > 0 {
			purgeConfirmText = purgeConfirmText[:len(purgeConfirmText)-1]
		}

		key := rl.GetCharPressed()
		for key > 0 {
			if key >= 32 && key <= 125 && len(purgeConfirmText) < 20 {
				purgeConfirmText += strings.ToUpper(string(rune(key)))
			}
			key = rl.GetCharPressed()
		}
	}

	textColor := rl.NewColor(255, 255, 255, 255)
	if purgeConfirmText == requiredPurgeText {
		textColor = rl.NewColor(100, 255, 100, 255)
	} else if len(purgeConfirmText) > 0 {
		textColor = rl.NewColor(255, 100, 100, 255)
	}

	inputTextX := modalX + 30
	inputTextY := inputY + 12
	rl.DrawText(purgeConfirmText, int32(inputTextX), int32(inputTextY), 16, textColor)

	if purgeTextActive && int(rl.GetTime()*2)%2 == 0 {
		cursorX := inputTextX + float32(rl.MeasureText(purgeConfirmText, 16))
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
		HideConfirmPurge()
		return
	}

	purgeWidth := float32(110)
	purgeRect := rl.NewRectangle(modalX+modalWidth-purgeWidth-20, buttonY, purgeWidth, buttonHeight)
	purgeHover := rl.CheckCollisionPointRec(mouse, purgeRect)
	canPurge := purgeConfirmText == requiredPurgeText

	var purgeBg rl.Color
	var purgeBorder rl.Color
	var purgeTextColor rl.Color

	if !canPurge {
		purgeBg = rl.NewColor(60, 30, 30, 255)
		purgeBorder = rl.NewColor(100, 50, 50, 255)
		purgeTextColor = rl.NewColor(150, 100, 100, 255)
	} else if purgeHover {
		purgeBg = rl.NewColor(255, 70, 70, 255)
		purgeBorder = rl.NewColor(255, 120, 120, 255)
		purgeTextColor = rl.NewColor(255, 255, 255, 255)
		glowPurge := rl.NewRectangle(purgeRect.X-1, purgeRect.Y-1, purgeRect.Width+2, purgeRect.Height+2)
		rl.DrawRectangleRounded(glowPurge, 0.2, 8, rl.NewColor(255, 70, 70, 80))
	} else {
		purgeBg = rl.NewColor(220, 50, 50, 255)
		purgeBorder = rl.NewColor(255, 80, 80, 255)
		purgeTextColor = rl.NewColor(255, 255, 255, 255)
	}

	rl.DrawRectangleRounded(purgeRect, 0.2, 6, purgeBg)
	rl.DrawRectangleRoundedLines(purgeRect, 0.2, 1, purgeBorder)
	rl.DrawText("Purge Item", int32(purgeRect.X+15), int32(purgeRect.Y+9), 16, purgeTextColor)

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && purgeHover && canPurge {
		start := time.Now()
		err := drivers.PurgeItem(purgeTargetName)
		status := "success"
		if err != nil {
			status = "failure"
			fmt.Printf("Purge failed: %v\n", err)
		}
		finished := time.Now()
		duration := int(finished.Sub(start).Seconds())

		var log WipeLog
		log.Wipe.Method = "secure_erase"
		log.Wipe.NistLevel = "purge"
		log.Wipe.Status = status
		log.Wipe.StartedAt = start.UTC().Format(time.RFC3339)
		log.Wipe.FinishedAt = finished.UTC().Format(time.RFC3339)
		log.Wipe.DurationSec = duration
		log.System.ToolVersion = "v1.0"
		log.System.HostOS = "Ubuntu 22.04"
		log.System.ExecutedBy = os.Getenv("USER")

		isDevice := strings.HasPrefix(purgeTargetName, "/dev/")
		if isDevice {
			devInfo, err := getDeviceInfo(purgeTargetName)
			if err == nil {
				log.Device.Name = devInfo["name"].(string)
				log.Device.Serial = devInfo["serial"].(string)
				log.Device.SizeGB = devInfo["size_gb"].(int)
				log.Device.Type = devInfo["type"].(string)
			}
		} else {
			fi, err := os.Stat(purgeTargetName)
			if err == nil {
				log.Device.Name = purgeTargetName
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
		HideConfirmPurge()
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		HideConfirmPurge()
	}
	if rl.IsKeyPressed(rl.KeyEnter) && canPurge {
		start := time.Now()
		err := drivers.PurgeItem(purgeTargetName)
		status := "success"
		if err != nil {
			status = "failure"
			fmt.Printf("Purge failed: %v\n", err)
		}
		finished := time.Now()
		duration := int(finished.Sub(start).Seconds())

		var log WipeLog
		log.Wipe.Method = "secure_erase"
		log.Wipe.NistLevel = "purge"
		log.Wipe.Status = status
		log.Wipe.StartedAt = start.UTC().Format(time.RFC3339)
		log.Wipe.FinishedAt = finished.UTC().Format(time.RFC3339)
		log.Wipe.DurationSec = duration
		log.System.ToolVersion = "v1.0"
		log.System.HostOS = "Ubuntu 22.04"
		log.System.ExecutedBy = os.Getenv("USER")

		isDevice := strings.HasPrefix(purgeTargetName, "/dev/")
		if isDevice {
			devInfo, err := getDeviceInfo(purgeTargetName)
			if err == nil {
				log.Device.Name = devInfo["name"].(string)
				log.Device.Serial = devInfo["serial"].(string)
				log.Device.SizeGB = devInfo["size_gb"].(int)
				log.Device.Type = devInfo["type"].(string)
			}
		} else {
			fi, err := os.Stat(purgeTargetName)
			if err == nil {
				log.Device.Name = purgeTargetName
				log.Device.Serial = ""
				log.Device.SizeGB = int(fi.Size() / 1000000000)
				log.Device.Type = "file"
			}
		}

		// Sign the log
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
		HideConfirmPurge()
	}
}

func getDeviceInfo(path string) (map[string]interface{}, error) {
	cmd := exec.Command("udevadm", "info", "--query=property", "--name", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	info := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			info[parts[0]] = parts[1]
		}
	}

	sizeCmd := exec.Command("lsblk", "-b", "-o", "SIZE", "-n", path)
	sizeOut, err := sizeCmd.Output()
	if err != nil {
		return nil, err
	}
	sizeBytes, _ := strconv.ParseInt(strings.TrimSpace(string(sizeOut)), 10, 64)
	sizeGB := int(sizeBytes / 1000000000)

	rotaCmd := exec.Command("lsblk", "-o", "ROTA", "-n", path)
	rotaOut, err := rotaCmd.Output()
	if err != nil {
		return nil, err
	}
	rotaStr := strings.TrimSpace(string(rotaOut))
	rotaInt, _ := strconv.Atoi(rotaStr)
	typ := "ssd"
	if rotaInt == 1 {
		typ = "hdd"
	}

	return map[string]interface{}{
		"name":    info["ID_MODEL"],
		"serial":  info["ID_SERIAL_SHORT"],
		"size_gb": sizeGB,
		"type":    typ,
	}, nil
}


