package progress

import (
	"fmt"
	"strconv"
	"strings"
)

// ShowProgressBar displays a progress bar based on percentage
func ShowProgressBar(percentage int) {
	const barWidth = 50
	filled := int(float64(percentage) * float64(barWidth) / 100.0)

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else if i == filled && percentage < 100 {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	fmt.Printf("\rMuxing subtitle tracks... %s %3d%%", bar, percentage)
	if percentage == 100 {
		fmt.Println(" Complete!")
	}
}

// ParseProgressLine extracts percentage from mkvmerge progress output
func ParseProgressLine(line string) (int, bool) {
	// In GUI mode, progress lines look like: "#GUI#progress 45%"
	if strings.HasPrefix(line, "#GUI#progress ") && strings.HasSuffix(line, "%") {
		percentStr := strings.TrimPrefix(line, "#GUI#progress ")
		percentStr = strings.TrimSuffix(percentStr, "%")
		if percentage, err := strconv.Atoi(strings.TrimSpace(percentStr)); err == nil {
			return percentage, true
		}
	}
	return 0, false
}
