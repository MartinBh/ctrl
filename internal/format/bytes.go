package format

import "fmt"

const unit = 1024

var byteUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}

func Bytes(value uint64) string {
	size := float64(value)
	unitIndex := 0

	for size >= unit && unitIndex < len(byteUnits)-1 {
		size /= unit
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d %s", value, byteUnits[unitIndex])
	}
	if size >= 10 {
		return fmt.Sprintf("%.0f %s", size, byteUnits[unitIndex])
	}

	return fmt.Sprintf("%.1f %s", size, byteUnits[unitIndex])
}
