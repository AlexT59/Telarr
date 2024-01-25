package types

import (
	"context"
	"strconv"
	"time"
)

type DownloadingStatus struct {
	// Found is true if the film is found in the queue.
	Found bool
	// FilmId is the id of the film.
	FilmId int64

	// Status is the status of the film in the queue.
	Status string

	// Size is the size of the download.
	Size float64
	// SizeLeft is the size left to download.
	SizeLeft float64
	// EstimatedCompletionTime is the estimated completion time of the download.
	EstimatedCompletionTime time.Time

	// ErrorMsg is the error message of the download.
	ErrorMsg string
}

type DownloadingStatusMessage struct {
	// GoroutineContextCancel is the cancel function of the goroutine.
	GoroutineContextCancel context.CancelFunc

	// MessageId is the id of the message.
	MessageId int
	// FilmId is the id of the film.
	FilmId int64

	// Ticker is the ticker of the goroutine.
	Ticker *time.Ticker
}

func (d DownloadingStatus) PrintDownloadingStatus(refreshRateSec int64) string {
	str := "*Status*: " + printStatus(d.Status) + "\n"

	if d.ErrorMsg != "" {
		str += "*Error*: " + d.ErrorMsg + "\n"
	}

	// add progress with icon to the string
	str += "*Progress*: \n" + printProgress(d.Size, d.SizeLeft, d.EstimatedCompletionTime) + "\n"

	str += "*Remaining time*: " + printRemainingTime(d.EstimatedCompletionTime) + "\n"

	if refreshRateSec > 0 {
		str += "_Refreshing every " + strconv.FormatInt(refreshRateSec, 10) + "s_\n"
	} else {
		str += "_Not refreshing_\n"
	}

	str += "\n"
	str += "_last update: " + time.Now().Format("2006-01-02 15:04:05") + "_\n"
	str += "_movieId: " + strconv.Itoa(int(d.FilmId)) + "_"

	return str
}

func (d DownloadingStatus) IsImported() bool {
	return d.Status == "imported"
}

func printStatus(status string) string {
	switch status {
	case "downloading":
		return "🟡 Downloading"
	case "importPending":
		return "🟠 Import pending"
	case "importing":
		return "🟡 Importing"
	case "imported":
		return "🟢 Imported"
	case "failedPending":
		return "🔴 Failed pending"
	case "failed":
		return "🔴 Failed"
	default:
		return "🔴 Unknown"
	}
}

func printProgress(size float64, sizeLeft float64, estimatedCompletionTime time.Time) string {
	sizeDl := size - sizeLeft
	progress := sizeDl / size * 100
	secLeft := time.Until(estimatedCompletionTime).Seconds()

	str := ""

	// (downloaded) of (left) GB
	mbDl := float64(sizeDl) / 1024 / 1024
	if mbDl > 1000 {
		str += strconv.FormatFloat(mbDl/1024, 'f', 2, 64) + " GB"
	} else {
		str += strconv.FormatFloat(mbDl, 'f', 2, 64) + " MB"
	}
	str += " of " + strconv.FormatFloat(float64(size)/1024/1024/1024, 'f', 2, 64) + " GB\n"

	// progress bar (see https://en.wikipedia.org/wiki/Block_Elements)
	for i := 0; i < 10; i++ {
		if progress > float64(i*10) {
			if progress > float64(i*10)+7.0/8*10 {
				str += "\u2588"
			} else if progress > float64(i*10)+3.0/4*10 {
				str += "\u2589"
			} else if progress > float64(i*10)+5.0/8*10 {
				str += "\u258A"
			} else if progress > float64(i*10)+1.0/2*10 {
				str += "\u258B"
			} else if progress > float64(i*10)+3.0/8*10 {
				str += "\u258C"
			} else if progress > float64(i*10)+1.0/4*10 {
				str += "\u258D"
			} else if progress > float64(i*10)+1.0/8*10 {
				str += "\u258E"
			} else {
				str += "\u258F"
			}
		} else {
			str += "\uFF0F"
		}
	}

	// progress in %
	str += " " + strconv.FormatFloat(progress, 'f', 2, 64) + "%\n"

	// download speed
	downloadSpeed := sizeLeft / 1024 / secLeft

	if downloadSpeed <= 0 {
		str += "- kB/s"
	} else if downloadSpeed > 1024 {
		str += strconv.FormatFloat(downloadSpeed/1024, 'f', 2, 64) + " MB/s"
	} else {
		str += strconv.FormatFloat(downloadSpeed, 'f', 2, 64) + " kB/s"
	}

	return str
}

func printRemainingTime(estimatedCompletionTime time.Time) string {
	durLeft := time.Until(estimatedCompletionTime)

	if durLeft.Seconds() < 0 {
		return "0s"
	}

	// convert to hours, minutes and seconds
	h := int(durLeft.Hours())
	m := int(durLeft.Minutes()) - h*60
	s := int(durLeft.Seconds()) - m*60 - h*60*60

	// format the string
	str := ""
	if h > 0 {
		str += strconv.Itoa(h) + "h"
	}
	if m > 0 {
		str += strconv.Itoa(m) + "m"
	}
	if s > 0 {
		str += strconv.Itoa(s) + "s"
	}

	return str
}
