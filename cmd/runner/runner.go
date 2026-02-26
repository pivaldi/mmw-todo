package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rivo/tview"
)

// Custom Writer to intercept and count lines
type progressWriter struct {
	target io.Writer
	onLine func()
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	lines := bytes.Count(p, []byte("\n"))
	for i := 0; i < lines; i++ {
		pw.onLine()
	}

	return pw.target.Write(p)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}

	return s
}

// --- 1. The Dynamic "Step-Back" Bar Engine ---
func buildIndeterminateBar(currentLen float64, maxWidth int) string {
	fractions := []string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

	filledLen := int(currentLen)
	if filledLen >= maxWidth {
		filledLen = maxWidth
	}

	// Calculate the sub-character resolution
	fractionIdx := int((currentLen - float64(filledLen)) * 8)
	if fractionIdx < 0 {
		fractionIdx = 0
	}
	if fractionIdx > 7 {
		fractionIdx = 7
	}

	filledStr := strings.Repeat("█", filledLen)
	fractionStr := ""
	if filledLen < maxWidth {
		fractionStr = fractions[fractionIdx]
	}

	emptyLen := maxWidth - filledLen
	if emptyLen > 0 && fractionStr != "" {
		emptyLen--
	}
	if emptyLen < 0 {
		emptyLen = 0
	}

	emptyStr := strings.Repeat("─", emptyLen)

	// Return the Hex-colored string
	return fmt.Sprintf("[#00E5FF]%s[#0088FF]%s[#444444]%s", filledStr, fractionStr, emptyStr)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go \"<command 1>\" \"<command 2>\"")
		os.Exit(1)
	}

	cmd1Str := os.Args[1]
	cmd2Str := os.Args[2]

	app := tview.NewApplication()

	// --- 2. Setup Panes ---
	left := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	left.SetBorder(true).SetTitle(fmt.Sprintf(" %s ", truncate(cmd1Str, 30))).SetBorderColor(tview.Styles.PrimaryTextColor)
	left.SetChangedFunc(func() {
		left.ScrollToEnd()
		app.Draw()
	})

	right := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	right.SetBorder(true).SetTitle(fmt.Sprintf(" %s ", truncate(cmd2Str, 30))).SetBorderColor(tview.Styles.PrimaryTextColor)
	right.SetChangedFunc(func() {
		right.ScrollToEnd()
		app.Draw()
	})

	status := tview.NewTextView().SetDynamicColors(true)

	// Layout
	topPanes := tview.NewFlex().
		AddItem(left, 0, 1, false).
		AddItem(right, 0, 1, false)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topPanes, 0, 1, false).
		AddItem(status, 1, 0, false)

	// --- 3. Animation State ---
	var currentLines int32 = 0
	done := make(chan struct{})

	updateProgress := func() {
		atomic.AddInt32(&currentLines, 1)
	}

	// Ticker for smooth animation updates
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond) // Faster tick for smoother bar movement
		defer ticker.Stop()

		spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frameIdx := 0

		var barProgress = 0.0
		var barSpeed = 0.5 // Characters to grow per tick

		for {
			select {
			case <-ticker.C:
				frameIdx = (frameIdx + 1) % len(spinnerFrames)
				spinChar := spinnerFrames[frameIdx]
				lines := atomic.LoadInt32(&currentLines)

				app.QueueUpdateDraw(func() {
					// A. Dynamically find the screen width
					_, _, screenWidth, _ := status.GetInnerRect()

					// B. Half the screen width
					halfScreen := float64(screenWidth) / 2.0
					if halfScreen < 10 {
						halfScreen = 10
					} // Sanity minimum

					// C. Grow the bar
					barProgress += barSpeed

					// D. The Magic "Step-Back" Logic!
					if barProgress >= halfScreen {
						// Drop back by 1/3 of its current width
						stepBackAmount := halfScreen / 3.0
						barProgress -= stepBackAmount
					}

					// E. Draw the UI
					bar := buildIndeterminateBar(barProgress, int(halfScreen))
					status.SetText(fmt.Sprintf(" Status: [#00E5FF]%s [white]%s [#0088FF]%d lines", spinChar, bar, lines))
				})

			case <-done:
				lines := atomic.LoadInt32(&currentLines)
				app.QueueUpdateDraw(func() {
					// When finished, fill the bar completely to half the screen in green
					_, _, screenWidth, _ := status.GetInnerRect()
					halfScreen := int(float64(screenWidth) / 2.0)
					finishedBar := fmt.Sprintf("[#00FF00]%s", strings.Repeat("█", halfScreen))

					status.SetText(fmt.Sprintf(" Status: [#00FF00]✓ [white]%s [#00FF00]Processed %d lines. Press Ctrl+C to exit.", finishedBar, lines))
				})

				return
			}
		}
	}()

	// --- 4. Command Runner ---
	var wg sync.WaitGroup
	wg.Add(2)

	runCommand := func(view *tview.TextView, cmdStr string) {
		defer wg.Done()
		cmd := exec.Command("sh", "-c", cmdStr)

		tracker := &progressWriter{
			target: tview.ANSIWriter(view),
			onLine: updateProgress,
		}
		cmd.Stdout = tracker
		cmd.Stderr = tracker

		cmd.Start()
		cmd.Wait()
		fmt.Fprintf(view, "\n[#00FF00]--- Complete ---[white]\n")
	}

	// --- 5. Launch Jobs ---
	go runCommand(left, cmd1Str)
	go runCommand(right, cmd2Str)

	go func() {
		wg.Wait()
		close(done)
	}()

	// --- 6. Start UI ---
	if err := app.SetRoot(mainLayout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// package main

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"os/exec"
// 	"strings"
// 	"sync"
// 	"sync/atomic"
// 	"time"

// 	"github.com/rivo/tview"
// )

// // --- 1. The Beautiful Progress Bar Engine ---
// func buildProgressBar(progress float64, width int) string {
// 	fractions := []string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// 	filledFloat := progress * float64(width)
// 	filledLen := int(filledFloat)

// 	if filledLen >= width {
// 		return fmt.Sprintf("[#00FF00]%s", strings.Repeat("█", width))
// 	}

// 	fractionIdx := int((filledFloat - float64(filledLen)) * 8)
// 	filledStr := strings.Repeat("█", filledLen)
// 	fractionStr := fractions[fractionIdx]

// 	emptyLen := width - filledLen
// 	if fractionIdx > 0 {
// 		emptyLen--
// 	}
// 	emptyStr := strings.Repeat("─", emptyLen)

// 	return fmt.Sprintf("[#00E5FF]%s[#0088FF]%s[#444444]%s", filledStr, fractionStr, emptyStr)
// }

// // --- 2. Custom Writer to Track Progress ---
// type progressWriter struct {
// 	target io.Writer
// 	onLine func()
// }

// func (pw *progressWriter) Write(p []byte) (int, error) {
// 	lines := bytes.Count(p, []byte("\n"))
// 	for i := 0; i < lines; i++ {
// 		pw.onLine()
// 	}
// 	return pw.target.Write(p)
// }

// func main() {
// 	app := tview.NewApplication()

// 	// --- Setup Panes ---
// 	left := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
// 	left.SetBorder(true).SetTitle(" Worker 1: Build Logs ").SetBorderColor(tview.Styles.PrimaryTextColor)
// 	left.SetChangedFunc(func() { app.Draw() })

// 	right := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
// 	right.SetBorder(true).SetTitle(" Worker 2: Gosec / Lint ").SetBorderColor(tview.Styles.PrimaryTextColor)
// 	right.SetChangedFunc(func() { app.Draw() })

// 	status := tview.NewTextView().SetDynamicColors(true)

// 	// --- Layout ---
// 	topPanes := tview.NewFlex().
// 		AddItem(left, 0, 1, false).
// 		AddItem(right, 0, 1, false)

// 	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
// 		AddItem(topPanes, 0, 1, false).
// 		AddItem(status, 1, 0, false)

// 	// --- Progress & Spinner State ---
// 	var totalSteps int32 = 25
// 	var currentSteps int32 = 0
// 	done := make(chan struct{}) // Channel to signal when everything is finished

// 	// The writer now ONLY increments the counter (thread-safe)
// 	updateProgress := func() {
// 		atomic.AddInt32(&currentSteps, 1)
// 	}

// 	// --- The Heartbeat (Ticker Goroutine) ---
// 	go func() {
// 		ticker := time.NewTicker(100 * time.Millisecond)
// 		defer ticker.Stop()

// 		spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
// 		frameIdx := 0

// 		for {
// 			select {
// 			case <-ticker.C:
// 				// 1. Advance the spinner
// 				frameIdx = (frameIdx + 1) % len(spinnerFrames)
// 				spinChar := spinnerFrames[frameIdx]

// 				// 2. Read the current progress safely
// 				curr := atomic.LoadInt32(&currentSteps)
// 				if curr > totalSteps {
// 					curr = totalSteps
// 				}

// 				// 3. Draw the UI
// 				app.QueueUpdateDraw(func() {
// 					progressRatio := float64(curr) / float64(totalSteps)
// 					percent := int(progressRatio * 100)
// 					bar := buildProgressBar(progressRatio, 40)

// 					// Notice the spinner [yellow]%s injected before the bar!
// 					status.SetText(fmt.Sprintf(" Status: [#00E5FF]%s [white]%s [white]%3d%%", spinChar, bar, percent))
// 				})

// 			case <-done:
// 				// When finished, replace the spinner with a green checkmark
// 				app.QueueUpdateDraw(func() {
// 					bar := buildProgressBar(1.0, 40)
// 					status.SetText(fmt.Sprintf(" Status: [#00FF00]✓ [white]%s [#00FF00]100%% (Finished! Press Ctrl+C)[white]", bar))
// 				})
// 				return // Kill the ticker goroutine
// 			}
// 		}
// 	}()

// 	// --- Command Runner ---
// 	var wg sync.WaitGroup
// 	wg.Add(2)

// 	runCommand := func(view *tview.TextView, cmdStr string) {
// 		defer wg.Done()
// 		cmd := exec.Command("sh", "-c", cmdStr)

// 		tracker := &progressWriter{
// 			target: tview.ANSIWriter(view),
// 			onLine: updateProgress,
// 		}
// 		cmd.Stdout = tracker
// 		cmd.Stderr = tracker

// 		cmd.Start()
// 		cmd.Wait()
// 		fmt.Fprintf(view, "\n[#00FF00]--- Complete ---[white]\n")
// 	}

// 	// --- Launch Jobs ---
// 	// cmd2 now has longer sleep gaps to prove the spinner keeps moving even when logs pause!
// 	cmd1 := `for i in $(seq 1 150); do echo -e "\033[36m[INFO]\033[0m Compiling pkg $i..."; sleep 0.02; done`
// 	cmd2 := `for i in $(seq 1 100); do echo -e "\033[33m[WARN]\033[0m Linting files $i..."; sleep 0.1; done`

// 	go runCommand(left, cmd1)
// 	go runCommand(right, cmd2)

// 	// Watcher for completion
// 	go func() {
// 		wg.Wait()
// 		close(done) // Triggers the 'case <-done:' in our Ticker select statement
// 	}()

// 	// --- Start UI ---
// 	if err := app.SetRoot(mainLayout, true).EnableMouse(true).Run(); err != nil {
// 		panic(err)
// 	}
// }
