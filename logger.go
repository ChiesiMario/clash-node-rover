package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

var (
	colorHeader  = color.New(color.FgHiMagenta, color.Bold)
	colorInfo    = color.New(color.FgCyan)
	colorSuccess = color.New(color.FgHiGreen, color.Bold)
	colorWarning = color.New(color.FgHiYellow)
	colorError   = color.New(color.FgHiRed, color.Bold)
	colorFailover= color.New(color.BgRed, color.FgHiWhite, color.Bold)
	colorMuted   = color.New(color.FgHiBlack)
	colorGroup   = color.New(color.FgHiBlue, color.Bold)
	colorNode    = color.New(color.FgHiWhite, color.Bold)
	colorValue   = color.New(color.FgHiGreen)
)

func logHeader(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(color.Output)
	log.Print(colorHeader.Sprintf("========== %s ==========", msg))
}

func logInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorInfo.Sprint("💡 ") + msg)
}

func logSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorSuccess.Sprint("✅ ") + msg)
}

func logWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorWarning.Sprint("⚠️ ") + msg)
}

func logError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorError.Sprint("❌ ") + msg)
}

func logFailover(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorFailover.Sprintf(" 🚨 急救機制 ") + " " + colorError.Sprintf(msg))
}

func logMuted(format string, a ...interface{}) {
	log.Print(colorMuted.Sprintf(format, a...))
}

func logGroup(group, format string, a ...interface{}) {
	prefix := colorGroup.Sprintf("[%s] ", group)
	msg := fmt.Sprintf(format, a...)
	log.Print(prefix + msg)
}

func formatNode(name string) string {
	return colorNode.Sprintf("[%s]", name)
}

func formatVal(val interface{}) string {
	return colorValue.Sprintf("%v", val)
}

func logReportStart() {
	timeStr := time.Now().Format("15:04:05")
	fmt.Fprintln(color.Output)
	fmt.Fprintln(color.Output, colorHeader.Sprintf("========== 週期測速報告 (%s) ==========", timeStr))
}

func logReportEnd() {
	fmt.Fprintln(color.Output, colorHeader.Sprint("======================================================="))
	fmt.Fprintln(color.Output)
}

func logGroupTitle(group string) {
	fmt.Fprintln(color.Output)
	fmt.Fprintln(color.Output, colorGroup.Sprintf("[%s]", group))
}

func logTreeItem(isLast bool, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	prefix := "  ├─ "
	if isLast {
		prefix = "  └─ "
	}
	fmt.Fprintln(color.Output, colorMuted.Sprint(prefix)+msg)
}
