package main

import (
	"time"
	"fmt"
)

// durationString converts duration to XyXmXd string
func durationString(d time.Duration) string {
    const dayHours = 24
    const monthHours = dayHours * 30
    const yearHours = monthHours * 12

    h := int(d.Hours())

    years := h / yearHours
    h = h % yearHours

    months := h / monthHours
    h = h % monthHours

    days := h / dayHours

    out := ""
    if years > 0 {
        out += fmt.Sprintf("%dy", years)
    }
    if months > 0 {
        out += fmt.Sprintf("%dm", months)
    }
    if days > 0 {
        out += fmt.Sprintf("%dd", days)
    }

    return out
}
