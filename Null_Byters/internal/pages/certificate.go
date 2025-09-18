package pages

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"math"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

type WipeLog struct {
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
	Signature struct {
		Algorithm            string `json:"algorithm"`
		Sig                  string `json:"sig"`
		PublicKeyFingerprint string `json:"public_key_fingerprint"`
		LogHash              string `json:"log_hash"`
	} `json:"signature"`
}

var (
	certificateActive        bool = false
	certificateLog           WipeLog
	certificateAnimationTime float32 = 0
	certificateScrollOffset  float32 = 0
	qrTexture                rl.Texture2D
	privateKey               ed25519.PrivateKey
	publicKey                ed25519.PublicKey
)

func init() {
	var err error
	publicKey, privateKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("Failed to generate key: %v\n", err)
	}
}

func ShowCertificate(log WipeLog) {
	certificateActive = true
	certificateLog = log
	certificateAnimationTime = 0
	certificateScrollOffset = 0

	// Generate QR code
	jsonBytes, _ := json.Marshal(log)
	qr, err := qrcode.New(string(jsonBytes), qrcode.Medium)
	if err == nil {
		qrImg := qr.Image(256)

		// Convert to Raylib Image
		rlImg := rl.NewImageFromImage(qrImg)
		qrTexture = rl.LoadTextureFromImage(rlImg)
		rl.UnloadImage(rlImg)
	}
}

func HideCertificate() {
	certificateActive = false
	certificateAnimationTime = 0
	certificateScrollOffset = 0
	if qrTexture.ID > 0 {
		rl.UnloadTexture(qrTexture)
	}
}

func IsCertificateActive() bool {
	return certificateActive
}

func DrawCertificate() {
	if !certificateActive {
		return
	}

	certificateAnimationTime += rl.GetFrameTime()

	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	overlayAlpha := uint8(math.Min(180, float64(certificateAnimationTime*300)))
	rl.DrawRectangle(0, 0, int32(screenWidth), int32(screenHeight),
		rl.NewColor(0, 0, 0, overlayAlpha))

	modalWidth := float32(600)
	modalHeight := float32(650) 
	modalX := (screenWidth - modalWidth) / 2
	modalY := (screenHeight - modalHeight) / 2

	scale := float32(math.Min(1.0, float64(certificateAnimationTime*4)))
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

	rl.DrawRectangleRoundedLines(modalRect, 0.15, 8, rl.NewColor(0, 255, 180, 255))

	glowRect := rl.NewRectangle(actualX-2, actualY-2, actualWidth+4, actualHeight+4)
	rl.DrawRectangleRounded(glowRect, 0.15, 12, rl.NewColor(0, 255, 180, 30))

	if scale < 1.0 {
		return
	}

	// Header
	headerHeight := float32(60)
	headerRect := rl.NewRectangle(modalX, modalY, modalWidth, headerHeight)
	rl.DrawRectangleRounded(headerRect, 0.15, 8, rl.NewColor(25, 35, 45, 200))

	iconSize := float32(24)
	iconX := modalX + 20
	iconY := modalY + (headerHeight-iconSize)/2

	rl.DrawCircle(int32(iconX+iconSize/2), int32(iconY+iconSize/2), iconSize/2, rl.NewColor(0, 255, 180, 255))
	rl.DrawCircleLines(int32(iconX+iconSize/2), int32(iconY+iconSize/2), iconSize/2, rl.NewColor(50, 255, 200, 255))

	rl.DrawText("i", int32(iconX+iconSize/2-3), int32(iconY+4), 16, rl.NewColor(5, 15, 20, 255))

	titleText := "Wipe Certificate"
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
		HideCertificate()
		return
	}

	
	contentRect := rl.NewRectangle(modalX+20, modalY+headerHeight+20, modalWidth-40, modalHeight-headerHeight-100)
	rl.BeginScissorMode(int32(contentRect.X), int32(contentRect.Y), int32(contentRect.Width), int32(contentRect.Height))

	contentY := float32(0) + certificateScrollOffset
	textColor := rl.NewColor(200, 200, 200, 255)
	labelColor := rl.NewColor(0, 255, 180, 255)

	// Device section
	rl.DrawTextEx(rl.GetFontDefault(), "Device:", rl.NewVector2(modalX+20, modalY+headerHeight+20+contentY), 16, 1, labelColor)
	contentY += 25
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Name: %s", certificateLog.Device.Name), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Serial: %s", certificateLog.Device.Serial), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Size: %d GB", certificateLog.Device.SizeGB), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Type: %s", certificateLog.Device.Type), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 30

	// Wipe section
	rl.DrawTextEx(rl.GetFontDefault(), "Wipe:", rl.NewVector2(modalX+20, modalY+headerHeight+20+contentY), 16, 1, labelColor)
	contentY += 25
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Method: %s", certificateLog.Wipe.Method), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("NIST Level: %s", certificateLog.Wipe.NistLevel), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Status: %s", certificateLog.Wipe.Status), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Started: %s", certificateLog.Wipe.StartedAt), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Finished: %s", certificateLog.Wipe.FinishedAt), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Duration: %d sec", certificateLog.Wipe.DurationSec), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 30

	// System section
	rl.DrawTextEx(rl.GetFontDefault(), "System:", rl.NewVector2(modalX+20, modalY+headerHeight+20+contentY), 16, 1, labelColor)
	contentY += 25
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Tool Version: %s", certificateLog.System.ToolVersion), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Host OS: %s", certificateLog.System.HostOS), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Executed By: %s", certificateLog.System.ExecutedBy), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 30

	// Signature section
	rl.DrawTextEx(rl.GetFontDefault(), "Signature:", rl.NewVector2(modalX+20, modalY+headerHeight+20+contentY), 16, 1, labelColor)
	contentY += 25
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Algorithm: %s", certificateLog.Signature.Algorithm), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	sigDisplay := certificateLog.Signature.Sig
	if len(sigDisplay) > 50 {
		sigDisplay = sigDisplay[:50] + "..."
	}
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Sig: %s", sigDisplay), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	fingerprintDisplay := certificateLog.Signature.PublicKeyFingerprint
	if len(fingerprintDisplay) > 50 {
		fingerprintDisplay = fingerprintDisplay[:50] + "..."
	}
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Public Key Fingerprint: %s", fingerprintDisplay), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 20
	rl.DrawTextEx(rl.GetFontDefault(), fmt.Sprintf("Log Hash (SHA256): %s", certificateLog.Signature.LogHash), rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, textColor)
	contentY += 30

	// QR Code
	rl.DrawTextEx(rl.GetFontDefault(), "QR Code (Scan for Log):", rl.NewVector2(modalX+20, modalY+headerHeight+20+contentY), 16, 1, labelColor)
	contentY += 25
	if qrTexture.ID > 0 {
		rl.DrawTexture(qrTexture, int32(modalX+40), int32(modalY+headerHeight+20+contentY), rl.White)
		contentY += float32(qrTexture.Height) + 20
	} else {
		rl.DrawTextEx(rl.GetFontDefault(), "QR Code generation failed", rl.NewVector2(modalX+40, modalY+headerHeight+20+contentY), 14, 1, rl.Red)
		contentY += 20
	}

	totalContentHeight := contentY - certificateScrollOffset

	rl.EndScissorMode()

	
	if totalContentHeight > contentRect.Height {
		scrollBarWidth := float32(10)
		scrollBarX := contentRect.X + contentRect.Width - scrollBarWidth
		scrollBarHeight := contentRect.Height
		trackRect := rl.NewRectangle(scrollBarX, contentRect.Y, scrollBarWidth, scrollBarHeight)
		rl.DrawRectangleRounded(trackRect, 0.5, 4, rl.NewColor(20, 60, 40, 150))

		thumbHeight := (contentRect.Height / totalContentHeight) * scrollBarHeight
		thumbY := contentRect.Y + (-certificateScrollOffset/totalContentHeight)*scrollBarHeight
		thumbRect := rl.NewRectangle(scrollBarX, thumbY, scrollBarWidth, thumbHeight)
		rl.DrawRectangleRounded(thumbRect, 0.5, 4, rl.NewColor(0, 255, 180, 255))

		// Handle scrolling
		wheelMove := rl.GetMouseWheelMove()
		if wheelMove != 0 {
			scrollSpeed := float32(50)
			certificateScrollOffset += wheelMove * scrollSpeed
			if certificateScrollOffset > 0 {
				certificateScrollOffset = 0
			}
			if certificateScrollOffset < -(totalContentHeight - contentRect.Height) {
				certificateScrollOffset = -(totalContentHeight - contentRect.Height)
			}
		}
	}

	// Buttons
	buttonY := modalY + modalHeight - 50
	buttonHeight := float32(35)
	exportWidth := float32(120)
	exportRect := rl.NewRectangle(modalX+modalWidth-exportWidth-120, buttonY, exportWidth, buttonHeight)
	exportHover := rl.CheckCollisionPointRec(mouse, exportRect)

	exportBg := rl.NewColor(0, 255, 180, 255)
	exportBorder := rl.NewColor(50, 255, 200, 255)
	exportTextColor := rl.NewColor(5, 15, 20, 255)
	if exportHover {
		exportBg = rl.NewColor(50, 255, 200, 255)
		exportBorder = rl.NewColor(100, 255, 220, 255)
		glowExport := rl.NewRectangle(exportRect.X-1, exportRect.Y-1, exportRect.Width+2, exportRect.Height+2)
		rl.DrawRectangleRounded(glowExport, 0.2, 8, rl.NewColor(50, 255, 200, 80))
	}

	rl.DrawRectangleRounded(exportRect, 0.2, 6, exportBg)
	rl.DrawRectangleRoundedLines(exportRect, 0.2, 1, exportBorder)
	rl.DrawText("Export PDF", int32(exportRect.X+15), int32(exportRect.Y+9), 16, exportTextColor)

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && exportHover {
		GeneratePDF(certificateLog)
	}

	closeWidth := float32(80)
	closeRectBtn := rl.NewRectangle(modalX+modalWidth-closeWidth-20, buttonY, closeWidth, buttonHeight)
	closeBtnHover := rl.CheckCollisionPointRec(mouse, closeRectBtn)

	closeBg := rl.NewColor(60, 60, 60, 255)
	closeBorder := rl.NewColor(120, 120, 120, 255)
	if closeBtnHover {
		closeBg = rl.NewColor(80, 80, 80, 255)
		closeBorder = rl.NewColor(160, 160, 160, 255)
	}

	rl.DrawRectangleRounded(closeRectBtn, 0.2, 6, closeBg)
	rl.DrawRectangleRoundedLines(closeRectBtn, 0.2, 1, closeBorder)
	rl.DrawText("Close", int32(closeRectBtn.X+20), int32(closeRectBtn.Y+9), 16, rl.NewColor(255, 255, 255, 255))

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && closeBtnHover {
		HideCertificate()
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		HideCertificate()
	}
}

func GeneratePDF(log WipeLog) {
	os.Mkdir("pdfs", 0755)


	jsonBytes, _ := json.Marshal(log)
	hash := sha256.Sum256(jsonBytes)
	log.Signature.LogHash = hex.EncodeToString(hash[:])

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	
	pdf.SetDrawColor(0, 150, 100)
	pdf.SetLineWidth(0.8)
	pdf.Rect(10, 10, 190, 277, "D")

	
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(0, 100, 80)
	pdf.CellFormat(0, 12, "Data Wipe Certificate", "", 1, "C", false, 0, "")
	pdf.Ln(4)
	pdf.SetFont("Arial", "I", 12)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 8, fmt.Sprintf("Generated: %s", log.Wipe.FinishedAt), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	
	sectionHeader := func(title string) {
		pdf.SetFillColor(230, 240, 235)
		pdf.SetTextColor(0, 100, 80)
		pdf.SetFont("Arial", "B", 14)
		pdf.CellFormat(0, 10, title, "1", 1, "L", true, 0, "")
		pdf.SetFont("Arial", "", 12)
		pdf.SetTextColor(0, 0, 0)
	}

	
	sectionHeader("Device Information")
	pdf.CellFormat(95, 8, fmt.Sprintf("Name: %s", log.Device.Name), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("Serial: %s", log.Device.Serial), "1", 1, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("Size: %d GB", log.Device.SizeGB), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("Type: %s", log.Device.Type), "1", 1, "L", false, 0, "")
	pdf.Ln(6)

	
	sectionHeader("Wipe Details")
	pdf.CellFormat(95, 8, fmt.Sprintf("Method: %s", log.Wipe.Method), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("NIST Level: %s", log.Wipe.NistLevel), "1", 1, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("Status: %s", log.Wipe.Status), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("Duration: %d sec", log.Wipe.DurationSec), "1", 1, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("Started: %s", log.Wipe.StartedAt), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("Finished: %s", log.Wipe.FinishedAt), "1", 1, "L", false, 0, "")
	pdf.Ln(6)

	
	sectionHeader("System Information")
	pdf.CellFormat(95, 8, fmt.Sprintf("Tool Version: %s", log.System.ToolVersion), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("Host OS: %s", log.System.HostOS), "1", 1, "L", false, 0, "")
	pdf.CellFormat(190, 8, fmt.Sprintf("Executed By: %s", log.System.ExecutedBy), "1", 1, "L", false, 0, "")
	pdf.Ln(6)

	
	sectionHeader("Signature & Verification")
	pdf.MultiCell(0, 8, fmt.Sprintf("Algorithm: %s", log.Signature.Algorithm), "1", "L", false)
	pdf.MultiCell(0, 8, fmt.Sprintf("Signature: %s", log.Signature.Sig), "1", "L", false)
	pdf.MultiCell(0, 8, fmt.Sprintf("Public Key Fingerprint: %s", log.Signature.PublicKeyFingerprint), "1", "L", false)
	pdf.MultiCell(0, 8, fmt.Sprintf("Log Hash (SHA256): %s", log.Signature.LogHash), "1", "L", false)
	pdf.Ln(6)

	
	sectionHeader("QR Code - Verification")
	qr, err := qrcode.New(string(jsonBytes), qrcode.Medium)
	if err == nil {
		qrImg := qr.Image(128)
		qrPath := "temp_qr.png"
		file, _ := os.Create(qrPath)
		png.Encode(file, qrImg)
		file.Close()
		pdf.Image(qrPath, 80, pdf.GetY()+5, 50, 50, false, "", 0, "")
		os.Remove(qrPath)
		pdf.Ln(60)
		pdf.SetFont("Arial", "I", 10)
		pdf.CellFormat(0, 8, "Scan this QR to verify wipe log authenticity", "", 1, "C", false, 0, "")
	}

	
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 10, "Confidential - Generated by Secure Wipe Tool", "", 0, "C", false, 0, "")

	
	fileName := fmt.Sprintf("pdfs/wipe_certificate_%s.pdf", strings.Replace(log.Wipe.StartedAt, ":", "-", -1))
	err = pdf.OutputFileAndClose(fileName)
	if err != nil {
		fmt.Printf("PDF generation failed: %v\n", err)
	} else {
		fmt.Printf("Saved PDF to %s\n", fileName)
	}
}
