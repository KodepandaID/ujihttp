package cli

import (
	"time"

	"github.com/gosuri/uitable"
	"github.com/i582/cfmt"
)

// DebugData for showing on terminal
type DebugData struct {
	Method     string
	Path       string
	Duration   time.Duration
	BodySize   int
	Code       int
	CodeStatus string
}

// WriteDebug to showing on terminal
func WriteDebug(d *DebugData) {
	table := uitable.New()
	table.MaxColWidth = 50

	table.AddRow(
		cfmt.Sprintf("{{%s}}::bold", "METHOD"),
		cfmt.Sprintf("{{%s}}::bold", "PATH"),
		cfmt.Sprintf("{{%s}}::bold", "SIZE"),
		cfmt.Sprintf("{{%s}}::bold", "DURATION"),
		cfmt.Sprintf("{{%s}}::bold", "StatusCode"))

	method := cfmt.Sprintf("{{%s}}::green|bold", d.Method)
	size := cfmt.Sprintf("{{%d B}}::bold", d.BodySize)

	durationCalculate := float32(d.Duration / time.Millisecond)
	duration := cfmt.Sprintf("{{%.2fms}}::green|bold", durationCalculate)
	if durationCalculate >= 200 && durationCalculate <= 1000 {
		duration = cfmt.Sprintf("{{%.2fms}}::red|bold", float32(d.Duration/time.Millisecond))
	}
	if durationCalculate >= 1000 {
		duration = cfmt.Sprintf("{{%.2fs}}::red|bold", float32(d.Duration/time.Second))
	}

	code := cfmt.Sprintf("{{%d %s}}::green|bold", d.Code, d.CodeStatus)
	if d.Code >= 300 {
		code = cfmt.Sprintf("{{%d %s}}::red|bold", d.Code, d.CodeStatus)
	}
	table.AddRow(method, d.Path, size, duration, code)
	cfmt.Println(table)
}
