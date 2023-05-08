package files

import (
	"fmt"
	"os"
)

func ByteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func GetLocalFileSize(fname string) (int64, error) {
	f1, err := os.Stat(fname)
	if err != nil {
		return 0, err
	}

	return f1.Size(), nil
}

// Create temp FIFO file
// func CreateTempFIFO()
