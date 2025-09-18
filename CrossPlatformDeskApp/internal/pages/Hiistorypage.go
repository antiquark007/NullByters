package pages

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var historyActive bool = false
var historyScrollOffset float32 = 0
var pdfFiles []string
var selectedPdf *string
var pdfDetails map[string]string 
var pdfSearchQuery string
var pdfSearchActive bool = false
var historyFileScrollOffset int

func drawHistoryTab() {
	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	
	if !historyActive {
		historyActive = true
		historyScrollOffset = 0
		if _, err := os.Stat("pdfs"); os.IsNotExist(err) {
			os.Mkdir("pdfs", 0755)
		}
		pdfFiles = nil
		pdfDetails = make(map[string]string)
		dir, err := os.ReadDir("pdfs")
		if err == nil {
			for _, entry := range dir {
				if !entry.IsDir() && strings.HasPrefix(entry.Name(), "wipe_certificate_") && strings.HasSuffix(entry.Name(), ".pdf") {
					pdfFiles = append(pdfFiles, entry.Name())
					pdfDetails[entry.Name()] = "Wipe Certificate"
				}
			}
		} else {
			fmt.Printf("Error reading pdfs directory: %v\n", err)
		}
	}

	const margin = 30.0
	const spacing = 12.0
	dialogsActive := IsConfirmPurgeActive() || IsConfirmClearActive() || IsCertificateActive()

	if selectedPdf == nil {
		totalPdfs := len(pdfFiles)
		if totalPdfs == 0 {
			
			
		
			messageBoxWidth := float32(400)
			messageBoxHeight := float32(200)
			messageBox := rl.NewRectangle(
				(screenWidth-messageBoxWidth)/2,
				(screenHeight-messageBoxHeight)/2,
				messageBoxWidth,
				messageBoxHeight,
			)
			
			rl.DrawRectangleRounded(messageBox, 0.2, 10, rl.NewColor(15, 60, 40, 200))
			rl.DrawRectangleRoundedLines(messageBox, 0.2, 10, rl.NewColor(0, 255, 180, 255))
			
		
			iconX := messageBox.X + messageBoxWidth/2 - 15
			iconY := messageBox.Y + 30
			rl.DrawText("📋", int32(iconX), int32(iconY), 48, rl.NewColor(0, 255, 180, 180))
			
			
			mainMsg := "No Certificates Available"
			mainMsgWidth := float32(rl.MeasureText(mainMsg, 24))
			mainMsgX := messageBox.X + (messageBoxWidth-mainMsgWidth)/2
			mainMsgY := messageBox.Y + 90
			rl.DrawText(mainMsg, int32(mainMsgX), int32(mainMsgY), 24, rl.NewColor(0, 255, 180, 255))
			
			
			subMsg := "No wipe certificates found in the pdfs folder."
			subMsgWidth := float32(rl.MeasureText(subMsg, 16))
			subMsgX := messageBox.X + (messageBoxWidth-subMsgWidth)/2
			subMsgY := messageBox.Y + 125
			rl.DrawText(subMsg, int32(subMsgX), int32(subMsgY), 16, rl.NewColor(0, 200, 150, 200))
			
			
			instrMsg := "Certificates will appear here after drive operations."
			instrMsgWidth := float32(rl.MeasureText(instrMsg, 14))
			instrMsgX := messageBox.X + (messageBoxWidth-instrMsgWidth)/2
			instrMsgY := messageBox.Y + 150
			rl.DrawText(instrMsg, int32(instrMsgX), int32(instrMsgY), 14, rl.NewColor(0, 180, 130, 180))
		} else {
			
			startY := float32(150.0)
			pdfHeight := float32(80.0)
			pdfSpacing := float32(90.0)

			
			availableHeight := screenHeight - startY - 60
			maxVisiblePdfs := int(availableHeight / pdfSpacing)
			if maxVisiblePdfs < 1 {
				maxVisiblePdfs = 1
			}

		
			if totalPdfs > maxVisiblePdfs {
				maxScrollPdfs := totalPdfs - maxVisiblePdfs
				wheelMove := rl.GetMouseWheelMove()

				if wheelMove < 0 && int(historyScrollOffset) < maxScrollPdfs {
					historyScrollOffset += 1
					if historyScrollOffset > float32(maxScrollPdfs) {
						historyScrollOffset = float32(maxScrollPdfs)
					}
				} else if wheelMove > 0 && historyScrollOffset > 0 {
					historyScrollOffset -= 1
				}

				if historyScrollOffset < 0 {
					historyScrollOffset = 0
				}
				if historyScrollOffset > float32(maxScrollPdfs) {
					historyScrollOffset = float32(maxScrollPdfs)
				}
			} else {
				historyScrollOffset = 0
			}

			
			clipRect := rl.NewRectangle(margin, startY, screenWidth-2*margin, availableHeight)
			rl.BeginScissorMode(int32(clipRect.X), int32(clipRect.Y), int32(clipRect.Width), int32(clipRect.Height))

			
			visibleCount := 0
			for i := int(historyScrollOffset); i < len(pdfFiles) && visibleCount < maxVisiblePdfs; i++ {
				pdf := pdfFiles[i]
				y := startY + float32(visibleCount)*pdfSpacing

				boxWidth := screenWidth - 2*margin - 15
				box := rl.NewRectangle(margin, y, boxWidth, pdfHeight)

				
				mouse := rl.GetMousePosition()
				hover := rl.CheckCollisionPointRec(mouse, box)

				if hover {
					rl.DrawRectangleRounded(box, 0.2, 10, rl.NewColor(15, 70, 45, 220))
					rl.DrawRectangleRoundedLines(box, 0.2, 10, rl.NewColor(50, 255, 200, 255))
				} else {
					rl.DrawRectangleRounded(box, 0.2, 10, rl.NewColor(10, 50, 30, 200))
					rl.DrawRectangleRoundedLines(box, 0.2, 10, rl.NewColor(0, 255, 180, 255))
				}

				
				rl.DrawText("📜", int32(box.X+20), int32(box.Y+25), 24, rl.NewColor(0, 255, 180, 255))

			
				displayName := pdf
				if len(displayName) > 35 {
					displayName = displayName[:32] + "..."
				}
				rl.DrawText(displayName, int32(box.X+60), int32(box.Y+15), 20, rl.NewColor(0, 255, 180, 255))

			
				infoText := "Type: Certificate"
				if details, exists := pdfDetails[pdf]; exists {
					infoText = fmt.Sprintf("Type: %s", details)
				}
				rl.DrawText(infoText, int32(box.X+60), int32(box.Y+40), 14, rl.NewColor(0, 200, 150, 200))

				
				viewBtn := rl.NewRectangle(box.X+boxWidth-240, box.Y+25, 80, 30)
				deleteBtn := rl.NewRectangle(box.X+boxWidth-140, box.Y+25, 80, 30)

				
				viewHover := rl.CheckCollisionPointRec(mouse, viewBtn)
				viewBg := rl.NewColor(0, 180, 255, 255)
				viewBorder := rl.NewColor(50, 200, 255, 255)
				viewTextColor := rl.NewColor(255, 255, 255, 255)
				if viewHover {
					viewBg = rl.NewColor(50, 200, 255, 255)
					viewBorder = rl.NewColor(100, 220, 255, 255)
					glowView := rl.NewRectangle(viewBtn.X-1, viewBtn.Y-1, viewBtn.Width+2, viewBtn.Height+2)
					rl.DrawRectangleRounded(glowView, 0.2, 6, rl.NewColor(50, 200, 255, 80))
				}
				rl.DrawRectangleRounded(viewBtn, 0.2, 6, viewBg)
				rl.DrawRectangleRoundedLines(viewBtn, 0.2, 1, viewBorder)
				rl.DrawText("View", int32(viewBtn.X+20), int32(viewBtn.Y+8), 14, viewTextColor)

				if rl.IsMouseButtonPressed(rl.MouseLeftButton) && viewHover && !dialogsActive {
					absPath, _ := filepath.Abs("pdfs/" + pdf)
					openPDF(absPath)
				}

				
				deleteHover := rl.CheckCollisionPointRec(mouse, deleteBtn)
				deleteBg := rl.NewColor(200, 50, 50, 255)
				deleteBorder := rl.NewColor(255, 100, 100, 255)
				deleteTextColor := rl.NewColor(255, 255, 255, 255)
				if deleteHover {
					deleteBg = rl.NewColor(255, 100, 100, 255)
					deleteBorder = rl.NewColor(255, 150, 150, 255)
					glowDelete := rl.NewRectangle(deleteBtn.X-1, deleteBtn.Y-1, deleteBtn.Width+2, deleteBtn.Height+2)
					rl.DrawRectangleRounded(glowDelete, 0.2, 6, rl.NewColor(255, 100, 100, 80))
				}
				rl.DrawRectangleRounded(deleteBtn, 0.2, 6, deleteBg)
				rl.DrawRectangleRoundedLines(deleteBtn, 0.2, 1, deleteBorder)
				rl.DrawText("Delete", int32(deleteBtn.X+15), int32(deleteBtn.Y+8), 14, deleteTextColor)

				if rl.IsMouseButtonPressed(rl.MouseLeftButton) && deleteHover && !dialogsActive {
					absPath, _ := filepath.Abs("pdfs/" + pdf)
					if err := os.Remove(absPath); err == nil {
						pdfFiles = append(pdfFiles[:i], pdfFiles[i+1:]...)
						delete(pdfDetails, pdf)
						
						if int(historyScrollOffset) >= len(pdfFiles) && historyScrollOffset > 0 {
							historyScrollOffset -= 1
						}
					} else {
						fmt.Printf("Failed to delete PDF %s: %v\n", pdf, err)
					}
				}

				visibleCount++
			}

			rl.EndScissorMode()

			
			if totalPdfs > maxVisiblePdfs {
				scrollBarWidth := float32(8.0)
				scrollBarX := screenWidth - margin - scrollBarWidth
				scrollBarHeight := float32(maxVisiblePdfs) * pdfSpacing

				
				trackRect := rl.NewRectangle(scrollBarX, startY, scrollBarWidth, scrollBarHeight)
				rl.DrawRectangleRounded(trackRect, 0.5, 4, rl.NewColor(20, 60, 40, 150))

				
				thumbHeight := (float32(maxVisiblePdfs) / float32(totalPdfs)) * scrollBarHeight
				if thumbHeight < 30 {
					thumbHeight = 30
				}
				thumbY := startY + (historyScrollOffset/float32(totalPdfs-maxVisiblePdfs))*(scrollBarHeight-thumbHeight)
				thumbRect := rl.NewRectangle(scrollBarX, thumbY, scrollBarWidth, thumbHeight)

				
				mouse := rl.GetMousePosition()
				if rl.CheckCollisionPointRec(mouse, thumbRect) {
					rl.DrawRectangleRounded(thumbRect, 0.5, 4, rl.NewColor(50, 255, 200, 255))
				} else {
					rl.DrawRectangleRounded(thumbRect, 0.5, 4, rl.NewColor(0, 255, 180, 255))
				}

				
				countText := fmt.Sprintf("Showing %d-%d of %d certificates",
					int(historyScrollOffset)+1,
					min(int(historyScrollOffset)+maxVisiblePdfs, totalPdfs),
					totalPdfs)
				countY := startY + scrollBarHeight + 10
				rl.DrawText(countText, int32(margin), int32(countY), 14, rl.NewColor(0, 255, 180, 180))
			}
		}

		
		startYForBtn := float32(150.0)
		if totalPdfs == 0 {
			startYForBtn = float32(100.0)
		}
		refreshBtn := rl.NewRectangle(screenWidth-margin-100, startYForBtn-50, 100, 40)
		drawGlowingButton(refreshBtn, "🔄 Refresh", rl.NewColor(0, 180, 255, 255), rl.NewColor(255, 255, 255, 255))
		if !dialogsActive && rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), refreshBtn) {
			dir, err := os.ReadDir("pdfs")
			if err == nil {
				pdfFiles = nil
				pdfDetails = make(map[string]string)
				for _, entry := range dir {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "wipe_certificate_") && strings.HasSuffix(entry.Name(), ".pdf") {
						pdfFiles = append(pdfFiles, entry.Name())
						pdfDetails[entry.Name()] = "Wipe Certificate"
					}
				}
			}
			historyScrollOffset = 0
		}
	}

	
	if rl.IsKeyPressed(rl.KeyEscape) {
		if selectedPdf != nil {
			selectedPdf = nil
		} else {
			historyActive = false
		}
	}
}


func getCurrentDate() string {
	return "2024-03-15 14:30:22" 
}


func openPDF(filePath string) {
	fmt.Printf("Attempting to open PDF: %s\n", filePath)
	
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("PDF file does not exist: %s\n", filePath)
		return
	}
	
	var cmd *exec.Cmd
	osType := detectOS()
	fmt.Printf("Detected OS: %s\n", osType)
	
	
	switch osType {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	default:
		fmt.Printf("Unsupported operating system '%s' for opening PDF: %s\n", osType, filePath)
		tryAlternativeOpeners(filePath)
		return
	}
	
	fmt.Printf("Executing command: %v\n", cmd.Args)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Primary command failed: %v\n", err)
		
		tryAlternativeOpeners(filePath)
	} else {
		fmt.Printf("Successfully opened PDF with primary method\n")
	}
}


func tryAlternativeOpeners(filePath string) {
	fmt.Printf("Trying alternative PDF openers for: %s\n", filePath)
	
	alternatives := [][]string{}
	osType := detectOS()
	
	switch osType {
	case "linux":
		alternatives = [][]string{
			{"xdg-open", filePath},
			{"evince", filePath},
			{"okular", filePath},
			{"zathura", filePath},
			{"mupdf", filePath},
			{"firefox", filePath},
			{"google-chrome", filePath},
			{"chromium-browser", filePath},
			{"atril", filePath},
			{"qpdfview", filePath},
		}
	case "darwin":
		alternatives = [][]string{
			{"open", filePath},
			{"open", "-a", "Preview", filePath},
			{"open", "-a", "Adobe Acrobat Reader DC", filePath},
			{"open", "-a", "Adobe Reader", filePath},
			{"open", "-a", "Firefox", filePath},
			{"open", "-a", "Google Chrome", filePath},
		}
	case "windows":
		alternatives = [][]string{
			{"start", filePath},
			{"explorer", filePath},
			{"rundll32", "url.dll,FileProtocolHandler", filePath},
		}
	}
	
	for i, cmdArgs := range alternatives {
		if len(cmdArgs) > 0 {
			fmt.Printf("Trying alternative %d: %v\n", i+1, cmdArgs)
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			if err := cmd.Run(); err == nil {
				fmt.Printf("✓ Successfully opened PDF with: %s\n", cmdArgs[0])
				return
			} else {
				fmt.Printf("✗ Failed with %s: %v\n", cmdArgs[0], err)
			}
		}
	}
	
	fmt.Printf("❌ Could not find any suitable PDF viewer for: %s\n", filePath)
	fmt.Printf("💡 Please install a PDF viewer or open the file manually at: %s\n", filePath)
	
	
	showInstallSuggestions(osType)
}


func showInstallSuggestions(osType string) {
	switch osType {
	case "linux":
		fmt.Printf("📝 Try installing a PDF viewer:\n")
		fmt.Printf("   • Ubuntu/Debian: sudo apt install evince\n")
		fmt.Printf("   • Fedora/CentOS: sudo dnf install evince\n")
		fmt.Printf("   • Arch: sudo pacman -S evince\n")
	case "darwin":
		fmt.Printf("📝 PDF viewers should be available. Try installing:\n")
		fmt.Printf("   • Adobe Acrobat Reader from adobe.com\n")
		fmt.Printf("   • Preview should be built-in\n")
	case "windows":
		fmt.Printf("📝 Try installing a PDF viewer:\n")
		fmt.Printf("   • Adobe Acrobat Reader\n")
		fmt.Printf("   • Microsoft Edge (built-in)\n")
	}
}


func detectOS() string {
	
	if osEnv := strings.ToLower(os.Getenv("OS")); strings.Contains(osEnv, "windows") {
		return "windows"
	}
	
	if osType := strings.ToLower(os.Getenv("OSTYPE")); osType != "" {
		if strings.Contains(osType, "linux") {
			return "linux"
		}
		if strings.Contains(osType, "darwin") || strings.Contains(osType, "mac") {
			return "darwin"
		}
		if strings.Contains(osType, "windows") {
			return "windows"
		}
	}
	
	
	if _, err := os.Stat("C:\\Windows"); err == nil {
		return "windows"
	}
	if strings.HasSuffix(strings.ToLower(os.Args[0]), ".exe") {
		return "windows"
	}
	
	// Check for macOS-specific indicators
	if _, err := os.Stat("/System/Library/CoreServices/Finder.app"); err == nil {
		return "darwin"
	}
	if _, err := os.Stat("/Applications"); err == nil {
		return "darwin"
	}
	
	// Check for Linux-specific indicators
	if _, err := os.Stat("/proc/version"); err == nil {
		return "linux"
	}
	if _, err := os.Stat("/etc/os-release"); err == nil {
		return "linux"
	}
	if _, err := os.Stat("/usr/bin"); err == nil {
		return "linux"
	}
	
	// Try to read /proc/version for more info
	if data, err := os.ReadFile("/proc/version"); err == nil {
		version := strings.ToLower(string(data))
		if strings.Contains(version, "linux") {
			return "linux"
		}
	}
	
	fmt.Printf("Warning: Could not detect operating system, defaulting to windows\n")
	return "windows" // Default fallback for Windows since you're on Windows
}


func isWindows() bool {
	return detectOS() == "windows"
}

func isMacOS() bool {
	return detectOS() == "darwin"
}

func isLinux() bool {
	return detectOS() == "linux"
}