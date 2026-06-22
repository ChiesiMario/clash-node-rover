package main

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

var (
	colorHeader  = color.New(color.FgHiMagenta, color.Bold)
	colorInfo    = color.New(color.FgCyan)
	colorSuccess = color.New(color.FgHiGreen, color.Bold)
	colorWarning = color.New(color.FgHiYellow)
	colorError   = color.New(color.FgHiRed, color.Bold)
	colorMuted   = color.New(color.FgHiBlack)
	colorGroup   = color.New(color.FgHiBlue, color.Bold)
	colorNode    = color.New(color.FgHiWhite, color.Bold)
	colorValue   = color.New(color.FgHiGreen)
)

func logHeader(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorHeader.Sprintf("\n========== %s ==========", msg))
}

func logInfo(format string, a ...interface{}) {
	log.Print(colorInfo.Sprintf("ℹ "+format, a...))
}

func logSuccess(format string, a ...interface{}) {
	log.Print(colorSuccess.Sprintf("✔ "+format, a...))
}

func logWarning(format string, a ...interface{}) {
	log.Print(colorWarning.Sprintf("⚠ "+format, a...))
}

func logError(format string, a ...interface{}) {
	log.Print(colorError.Sprintf("✖ "+format, a...))
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
