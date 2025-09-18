package pages

import (
	"data_wiper/internal/drivers"
	"fmt"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var cachedDrives []drivers.Drive
var selectedDrive *drivers.Drive
var driveContents []string
var searchQuery string
var scrollOffset int
var driveScrollOffset int
var maxVisibleItems = 10
var maxVisibleDrives = 5
var searchActive bool = false

const (
	TabDrives = iota
	TabHistory
	TabSettings
)

var activeTab = TabDrives

var rotation float32 = 0

func DrawDashboard() {
    screenWidth := float32(rl.GetScreenWidth())
    screenHeight := float32(rl.GetScreenHeight())

   
    rl.DrawRectangleGradientV(
        0, 0, int32(screenWidth), int32(screenHeight),
        rl.NewColor(5, 15, 20, 255),
        rl.NewColor(0, 200, 120, 255),
    )

    for i := 0; i < 60; i++ {
        x := float32((i*47+int(rl.GetTime()*25))%int(screenWidth))
        y := float32((i*89+int(rl.GetTime()*35))%int(screenHeight))
        rl.DrawCircle(int32(x), int32(y), 2, rl.NewColor(0, 255, 180, 150))
    }

  
    titleWidth := float32(rl.MeasureText("Dashboard - Null Byters", 32))
    rl.DrawText("Dashboard - Null Byters", int32((screenWidth-titleWidth)/2), 20, 32, rl.NewColor(0, 255, 180, 255))

   
    tabWidth := float32(140)
    tabHeight := float32(40)
    tabY := float32(80) 
    tabSpacing := float32(12)
    tabs := []string{"Drives", "History"}
    totalTabsWidth := float32(len(tabs))*tabWidth + float32(len(tabs)-1)*tabSpacing
    tabStartX := (screenWidth - totalTabsWidth) / 2 

    for i, t := range tabs {
        tabX := tabStartX + float32(i)*(tabWidth+tabSpacing)
        rect := rl.NewRectangle(tabX, tabY, tabWidth, tabHeight)

        mouse := rl.GetMousePosition()
        hover := rl.CheckCollisionPointRec(mouse, rect)

        if activeTab == i {
            rl.DrawRectangleRounded(rect, 0.3, 8, rl.NewColor(0, 255, 180, 255))
            rl.DrawRectangleRoundedLines(rect, 0.3, 8, rl.NewColor(50, 255, 200, 255))
            rl.DrawText(t, int32(rect.X+(rect.Width-float32(rl.MeasureText(t, 20)))/2), int32(rect.Y+10), 20, rl.NewColor(5, 15, 20, 255))
        } else if hover {
            rl.DrawRectangleRounded(rect, 0.3, 8, rl.NewColor(20, 100, 60, 255))
            rl.DrawRectangleRoundedLines(rect, 0.3, 8,  rl.NewColor(0, 255, 180, 255))
            rl.DrawText(t, int32(rect.X+(rect.Width-float32(rl.MeasureText(t, 20)))/2), int32(rect.Y+10), 20, rl.NewColor(0, 255, 180, 255))
        } else {
            rl.DrawRectangleRounded(rect, 0.3, 8, rl.NewColor(10, 50, 30, 255))
            rl.DrawRectangleRoundedLines(rect, 0.3, 8,  rl.NewColor(0, 255, 180, 255))
            rl.DrawText(t, int32(rect.X+(rect.Width-float32(rl.MeasureText(t, 20)))/2), int32(rect.Y+10), 20, rl.NewColor(0, 255, 180, 255))
        }

        if rl.IsMouseButtonPressed(rl.MouseLeftButton) && hover {
            activeTab = i
        }
    }

    switch activeTab {
    case TabDrives:
        drawDrivesTab(screenWidth, screenHeight)
    case TabHistory:
        drawHistoryTab()
    case TabSettings:
        drawSettingsTab()
    }
    DrawConfirmClear()
    DrawConfirmPurge()
    if IsCertificateActive() {
        DrawCertificate()
    }
}

func drawDrivesTab(screenWidth, screenHeight float32) {
    if cachedDrives == nil {
        d, _ := drivers.GetDrives()
        cachedDrives = d
    }

    const margin = 30.0
    const spacing = 12.0
    dialogsActive := IsConfirmPurgeActive() || IsConfirmClearActive() || IsCertificateActive()

    if selectedDrive == nil {
        
        startY := float32(150.0) 
        driveHeight := float32(80.0)
        driveSpacing := float32(90.0)

       
        availableHeight := screenHeight - startY - 60
        maxVisibleDrives = int(availableHeight / driveSpacing)
        if maxVisibleDrives < 1 {
            maxVisibleDrives = 1
        }

        totalDrives := len(cachedDrives)

       
        if totalDrives > maxVisibleDrives {
            maxScrollDrives := totalDrives - maxVisibleDrives
            wheelMove := rl.GetMouseWheelMove()

            if wheelMove < 0 && driveScrollOffset < maxScrollDrives {
                driveScrollOffset++
                if driveScrollOffset > maxScrollDrives {
                    driveScrollOffset = maxScrollDrives
                }
            } else if wheelMove > 0 && driveScrollOffset > 0 {
                driveScrollOffset--
            }

            if driveScrollOffset < 0 {
                driveScrollOffset = 0
            }
            if driveScrollOffset > maxScrollDrives {
                driveScrollOffset = maxScrollDrives
            }
        } else {
            driveScrollOffset = 0
        }

        clipRect := rl.NewRectangle(margin, startY, screenWidth-2*margin, availableHeight)
        rl.BeginScissorMode(int32(clipRect.X), int32(clipRect.Y), int32(clipRect.Width), int32(clipRect.Height))

        visibleCount := 0
        for i := driveScrollOffset; i < len(cachedDrives) && visibleCount < maxVisibleDrives; i++ {
            d := cachedDrives[i]
            y := startY + float32(visibleCount)*driveSpacing

            boxWidth := screenWidth - 2*margin - 15
            box := rl.NewRectangle(margin, y, boxWidth, driveHeight)

          
            mouse := rl.GetMousePosition()
            hover := rl.CheckCollisionPointRec(mouse, box)

            if hover {
                rl.DrawRectangleRounded(box, 0.2, 10, rl.NewColor(15, 70, 45, 220))
                rl.DrawRectangleRoundedLines(box, 0.2, 10,  rl.NewColor(50, 255, 200, 255))
            } else {
                rl.DrawRectangleRounded(box, 0.2, 10, rl.NewColor(10, 50, 30, 200))
                rl.DrawRectangleRoundedLines(box, 0.2, 10,  rl.NewColor(0, 255, 180, 255))
            }

          
            iconText := "💾"
            if d.Type == "usb" {
                iconText = "🔌"
            } else if d.Type == "network" {
                iconText = "🌐"
            }
            rl.DrawText(iconText, int32(box.X+20), int32(box.Y+25), 24, rl.NewColor(0, 255, 180, 255))

          
            nameText := fmt.Sprintf("Drive: %s (%s)", d.Name, d.Device)
            rl.DrawText(nameText, int32(box.X+60), int32(box.Y+15), 20, rl.NewColor(0, 255, 180, 255))

            
            infoText := fmt.Sprintf("Type: %s", d.Type)
            if d.FileSystem != "" {
                infoText += fmt.Sprintf(" | FS: %s", d.FileSystem)
            }
            if d.IsRemovable {
                infoText += " | Removable"
            }
            rl.DrawText(infoText, int32(box.X+60), int32(box.Y+40), 14, rl.NewColor(0, 200, 150, 200))

            if !dialogsActive && rl.IsMouseButtonPressed(rl.MouseLeftButton) && hover {
                selectedDrive = &cachedDrives[i]
                driveContents, _ = drivers.GetDriveContents(selectedDrive.Path)
                searchQuery = ""
                searchActive = false
                scrollOffset = 0
            }

            visibleCount++
        }

        rl.EndScissorMode()

        
        if totalDrives > maxVisibleDrives {
            scrollBarWidth := float32(8.0)
            scrollBarX := screenWidth - margin - scrollBarWidth
            scrollBarHeight := float32(maxVisibleDrives) * driveSpacing

            
            trackRect := rl.NewRectangle(scrollBarX, startY, scrollBarWidth, scrollBarHeight)
            rl.DrawRectangleRounded(trackRect, 0.5, 4, rl.NewColor(20, 60, 40, 150))

            
            thumbHeight := (float32(maxVisibleDrives) / float32(totalDrives)) * scrollBarHeight
            if thumbHeight < 30 {
                thumbHeight = 30
            }
            thumbY := startY + (float32(driveScrollOffset)/float32(totalDrives-maxVisibleDrives))*(scrollBarHeight-thumbHeight)
            thumbRect := rl.NewRectangle(scrollBarX, thumbY, scrollBarWidth, thumbHeight)

            
            mouse := rl.GetMousePosition()
            if rl.CheckCollisionPointRec(mouse, thumbRect) {
                rl.DrawRectangleRounded(thumbRect, 0.5, 4, rl.NewColor(50, 255, 200, 255))
            } else {
                rl.DrawRectangleRounded(thumbRect, 0.5, 4, rl.NewColor(0, 255, 180, 255))
            }

           
            countText := fmt.Sprintf("Showing %d-%d of %d drives", 
                driveScrollOffset+1, 
                min(driveScrollOffset+maxVisibleDrives, totalDrives), 
                totalDrives)
            countY := startY + scrollBarHeight + 10
            rl.DrawText(countText, int32(margin), int32(countY), 14, rl.NewColor(0, 255, 180, 180))
        }

        
        refreshBtn := rl.NewRectangle(screenWidth-margin-100, startY-50, 100, 40)
        drawGlowingButton(refreshBtn, "🔄 Refresh", rl.NewColor(0, 180, 255, 255), rl.NewColor(255, 255, 255, 255))
        if !dialogsActive && rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), refreshBtn) {
            d, _ := drivers.GetDrives()
            cachedDrives = d
            driveScrollOffset = 0
        }
    } else {
        
        searchHeight := float32(50)
        buttonHeight := float32(40)
        headerY := float32(150)

        
        backBtn := rl.NewRectangle(margin, headerY, 100, buttonHeight)
        drawGlowingButton(backBtn, "⬅ Back", rl.NewColor(0, 255, 180, 255), rl.NewColor(5, 15, 20, 255))
        if !dialogsActive && rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), backBtn) {
            selectedDrive = nil
            searchQuery = ""
            searchActive = false
            scrollOffset = 0
            return
        }

     
        purgeDriveBtn := rl.NewRectangle(margin+110+spacing, headerY, 100, buttonHeight)
        drawGlowingButton(purgeDriveBtn, "Purge Drive", rl.NewColor(200, 50, 50, 255), rl.NewColor(255, 255, 255, 255))
        if !dialogsActive && rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), purgeDriveBtn) {
            ShowConfirmPurge(selectedDrive.Path)
        }

        clearDriveBtn := rl.NewRectangle(margin+220+2*spacing, headerY, 100, buttonHeight)
        drawGlowingButton(clearDriveBtn, "Clear Drive", rl.NewColor(0, 255, 180, 255), rl.NewColor(5, 15, 20, 255))
        if !dialogsActive && rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), clearDriveBtn) {
            ShowConfirmClear(selectedDrive.Path)
        }

       
        driveInfoRect := rl.NewRectangle(margin+330+3*spacing, headerY, 300, buttonHeight)
        rl.DrawRectangleRounded(driveInfoRect, 0.2, 6, rl.NewColor(15, 60, 40, 180))
        rl.DrawRectangleRoundedLines(driveInfoRect, 0.2, 6,  rl.NewColor(0, 255, 180, 255))
        infoText := fmt.Sprintf("Drive: %s (%s)", selectedDrive.Name, selectedDrive.Device)
        rl.DrawText(infoText, int32(driveInfoRect.X+12), int32(driveInfoRect.Y+10), 16, rl.NewColor(0, 255, 180, 255))

      
        searchY := headerY + buttonHeight + spacing + 10
        searchRect := rl.NewRectangle(margin, searchY, screenWidth-2*margin, searchHeight)

        mouse := rl.GetMousePosition()
        if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, searchRect) {
            searchActive = true
        } else if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !rl.CheckCollisionPointRec(mouse, searchRect) {
            searchActive = false
        }

        if searchActive && !dialogsActive {
            rl.DrawRectangleRounded(searchRect, 0.2, 8, rl.NewColor(25, 80, 50, 220))
            rl.DrawRectangleRoundedLines(searchRect, 0.2, 8, rl.NewColor(50, 255, 200, 255))
        } else {
            rl.DrawRectangleRounded(searchRect, 0.2, 8, rl.NewColor(15, 60, 40, 200))
            rl.DrawRectangleRoundedLines(searchRect, 0.2, 8, rl.NewColor(0, 255, 180, 255))
        }

        searchIconRect := rl.NewRectangle(searchRect.X+12, searchRect.Y+12, 80, 30)
        rl.DrawRectangleRounded(searchIconRect, 0.2, 4, rl.NewColor(0, 255, 180, 120))
        rl.DrawText("🔍 Search", int32(searchIconRect.X+10), int32(searchIconRect.Y+8), 16, rl.NewColor(0, 255, 180, 255))

        queryTextX := searchRect.X + 100
        queryTextY := searchRect.Y + 16
        if searchQuery == "" && !searchActive {
            rl.DrawText("Type to search files and folders...", int32(queryTextX), int32(queryTextY), 16, rl.NewColor(100, 200, 150, 180))
        } else {
            rl.DrawText(searchQuery, int32(queryTextX), int32(queryTextY), 16, rl.NewColor(255, 255, 255, 255))
            if searchActive && int(rl.GetTime()*2)%2 == 0 {
                cursorX := queryTextX + float32(rl.MeasureText(searchQuery, 16))
                rl.DrawText("|", int32(cursorX), int32(queryTextY), 16, rl.NewColor(0, 255, 180, 255))
            }
        }

        if searchActive && !dialogsActive {
            if rl.IsKeyPressed(rl.KeyBackspace) && len(searchQuery) > 0 {
                searchQuery = searchQuery[:len(searchQuery)-1]
            }

            key := rl.GetCharPressed()
            for key > 0 {
                if key >= 32 && key <= 125 && len(searchQuery) < 50 {
                    searchQuery += string(rune(key))
                }
                key = rl.GetCharPressed()
            }
        }

        filtered := []string{}
        if searchQuery == "" {
            filtered = driveContents
        } else {
            for _, f := range driveContents {
                if containsIgnoreCase(f, searchQuery) {
                    filtered = append(filtered, f)
                }
            }
        }

       
        fileListStartY := searchY + searchHeight + spacing + 10
        availableHeight := screenHeight - fileListStartY - 60
        itemHeight := float32(40.0)
        maxVisibleItems = int(availableHeight / itemHeight)
        if maxVisibleItems < 1 {
            maxVisibleItems = 1
        }

        totalItems := len(filtered)
        if totalItems > maxVisibleItems {
            maxScrollItems := totalItems - maxVisibleItems
            wheelMove := rl.GetMouseWheelMove()

            if wheelMove < 0 && scrollOffset < maxScrollItems {
                scrollOffset++
                if scrollOffset > maxScrollItems {
                    scrollOffset = maxScrollItems
                }
            } else if wheelMove > 0 && scrollOffset > 0 {
                scrollOffset--
            }
        } else {
            scrollOffset = 0
        }

        buttonWidth := float32(60)
        buttonSpacing := float32(8)
        fileListWidth := screenWidth - 2*margin - (2*buttonWidth + buttonSpacing + 15)

        visibleCount := 0
        for i := scrollOffset; i < len(filtered) && visibleCount < maxVisibleItems; i++ {
            f := filtered[i]
            y := fileListStartY + float32(visibleCount)*itemHeight

            fileRect := rl.NewRectangle(margin, y, fileListWidth, itemHeight-3)

            var fileColor rl.Color
            var bgColor rl.Color
            if searchQuery != "" && containsIgnoreCase(f, searchQuery) {
                bgColor = rl.NewColor(20, 80, 50, 200)
                fileColor = rl.NewColor(100, 255, 200, 255)
            } else {
                bgColor = rl.NewColor(10, 50, 30, 200)
                fileColor = rl.NewColor(0, 255, 180, 255)
            }

            rl.DrawRectangleRounded(fileRect, 0.2, 6, bgColor)
            rl.DrawRectangleRoundedLines(fileRect, 0.2, 6, rl.NewColor(0, 255, 180, 255))

            displayName := f
            maxChars := int(fileListWidth / 9)
            if len(f) > maxChars {
                displayName = f[:maxChars-3] + "..."
            }
            rl.DrawText(displayName, int32(fileRect.X+12), int32(fileRect.Y+10), 16, fileColor)

            purgeBtn := rl.NewRectangle(fileRect.X+fileListWidth+10, fileRect.Y+4, buttonWidth, itemHeight-8)
            clearBtn := rl.NewRectangle(purgeBtn.X+buttonWidth+buttonSpacing, fileRect.Y+4, buttonWidth, itemHeight-8)

            drawGlowingButton(purgeBtn, "Purge", rl.NewColor(200, 50, 50, 255), rl.NewColor(255, 255, 255, 255))
            drawGlowingButton(clearBtn, "Clear", rl.NewColor(0, 255, 180, 255), rl.NewColor(5, 15, 20, 255))

            if !dialogsActive && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
                mouse := rl.GetMousePosition()
                if rl.CheckCollisionPointRec(mouse, purgeBtn) {
                    ShowConfirmPurge(f)
                }
                if rl.CheckCollisionPointRec(mouse, clearBtn) {
                    ShowConfirmClear(f)
                }
            }

            visibleCount++
        }

        if totalItems > maxVisibleItems {
            scrollBarWidth := float32(8.0)
            scrollBarX := screenWidth - margin - scrollBarWidth
            scrollBarHeight := float32(maxVisibleItems) * itemHeight

            trackRect := rl.NewRectangle(scrollBarX, fileListStartY, scrollBarWidth, scrollBarHeight)
            rl.DrawRectangleRounded(trackRect, 0.5, 4, rl.NewColor(20, 60, 40, 150))

            thumbHeight := (float32(maxVisibleItems) / float32(totalItems)) * scrollBarHeight
            thumbY := fileListStartY + (float32(scrollOffset)/float32(totalItems-maxVisibleItems))*(scrollBarHeight-thumbHeight)
            thumbRect := rl.NewRectangle(scrollBarX, thumbY, scrollBarWidth, thumbHeight)
            rl.DrawRectangleRounded(thumbRect, 0.5, 4, rl.NewColor(0, 255, 180, 255))
        }

        if searchQuery != "" || len(filtered) < len(driveContents) {
            resultText := fmt.Sprintf("Showing %d of %d items", len(filtered), len(driveContents))
            if searchQuery != "" {
                resultText = fmt.Sprintf("Found %d results for \"%s\"", len(filtered), searchQuery)
            }
            resultY := fileListStartY + float32(maxVisibleItems)*itemHeight + 10
            rl.DrawText(resultText, int32(margin), int32(resultY), 14, rl.NewColor(0, 255, 180, 180))
        }
    }
}

func drawGlowingButton(rect rl.Rectangle, text string, bg rl.Color, fg rl.Color) {
    mouse := rl.GetMousePosition()
    hover := rl.CheckCollisionPointRec(mouse, rect)

    var buttonBg rl.Color
    var borderColor rl.Color

    if hover {
        glowRect := rl.NewRectangle(rect.X-2, rect.Y-2, rect.Width+4, rect.Height+4)
        rl.DrawRectangleRounded(glowRect, 0.3, 6, rl.NewColor(bg.R, bg.G, bg.B, 80))
        buttonBg = rl.NewColor(
            uint8(min(255, int(bg.R)+20)),
            uint8(min(255, int(bg.G)+20)),
            uint8(min(255, int(bg.B)+20)),
            bg.A,
        )
        borderColor = rl.NewColor(255, 255, 100, 200)
    } else {
        buttonBg = bg
        borderColor = rl.NewColor(50, 255, 200, 150)
    }

    rl.DrawRectangleRounded(rect, 0.3, 6, buttonBg)
    rl.DrawRectangleRoundedLines(rect, 0.3, 6,  borderColor)

    fontSize := 16
    textWidth := rl.MeasureText(text, int32(fontSize))
    textX := rect.X + (rect.Width-float32(textWidth))/2
    textY := rect.Y + (rect.Height-float32(fontSize))/2 + 2
    rl.DrawText(text, int32(textX), int32(textY), int32(fontSize), fg)
}

func containsIgnoreCase(str, substr string) bool {
    return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

