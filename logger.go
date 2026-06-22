package main

import (
	"fmt"
	"log"
	"sync"
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

type WebLogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

var (
	logHistory      = make([]WebLogEntry, 0, 50)
	logHistoryMutex sync.Mutex
)

func broadcastWebLog(level, msg string) {
	entry := WebLogEntry{
		Level:   level,
		Message: msg,
		Time:    time.Now().Format("15:04:05"),
	}

	logHistoryMutex.Lock()
	if len(logHistory) >= 50 {
		logHistory = logHistory[1:]
	}
	logHistory = append(logHistory, entry)
	logHistoryMutex.Unlock()

	BroadcastSingleLog(entry)
}

func logHeader(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(color.Output)
	log.Print(colorHeader.Sprintf("========== %s ==========", msg))
	broadcastWebLog("header", fmt.Sprintf("========== %s ==========", msg))
}

func logInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorInfo.Sprint("💡 ") + msg)
	broadcastWebLog("info", "💡 "+msg)
}

func logSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorSuccess.Sprint("✅ ") + msg)
	broadcastWebLog("success", "✅ "+msg)
}

func logWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorWarning.Sprint("⚠️ ") + msg)
	broadcastWebLog("warning", "⚠️ "+msg)
}

func logError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorError.Sprint("❌ ") + msg)
	broadcastWebLog("error", "❌ "+msg)
}

func logFailover(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorFailover.Sprintf(" 🚨 急救機制 ") + " " + colorError.Sprintf(msg))
	broadcastWebLog("error", "🚨 [急救機制] "+msg)
}

func logMuted(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(colorMuted.Sprintf(msg))
	broadcastWebLog("muted", msg)
}

func logGroup(group, format string, a ...interface{}) {
	prefix := colorGroup.Sprintf("[%s] ", group)
	msg := fmt.Sprintf(format, a...)
	log.Print(prefix + msg)
	broadcastWebLog("info", fmt.Sprintf("[%s] %s", group, msg))
}

var (
	GlobalNodeProviders = make(map[string]string)
	providerMutex       sync.RWMutex
)

func SetNodeProvider(name, provider string) {
	providerMutex.Lock()
	defer providerMutex.Unlock()
	GlobalNodeProviders[name] = provider
}

func GetNodeProvider(name string) string {
	providerMutex.RLock()
	defer providerMutex.RUnlock()
	return GlobalNodeProviders[name]
}

func formatNode(name string) string {
	provider := GetNodeProvider(name)
	if provider != "" {
		return colorNode.Sprintf("[%s: %s]", provider, name)
	}
	return colorNode.Sprintf("[%s]", name)
}

func formatVal(val interface{}) string {
	return colorValue.Sprintf("%v", val)
}

func logReportStart() {
	timeStr := time.Now().Format("15:04:05")
	fmt.Fprintln(color.Output)
	fmt.Fprintln(color.Output, colorHeader.Sprintf("========== 週期測速報告 (%s) ==========", timeStr))
	broadcastWebLog("header", fmt.Sprintf("========== 週期測速報告 (%s) ==========", timeStr))
}

func logReportEnd() {
	fmt.Fprintln(color.Output, colorHeader.Sprint("======================================================="))
	fmt.Fprintln(color.Output)
	broadcastWebLog("header", "=======================================================")
}

func logGroupTitle(group string) {
	fmt.Fprintln(color.Output)
	fmt.Fprintln(color.Output, colorGroup.Sprintf("[%s]", group))
	broadcastWebLog("group", fmt.Sprintf("[%s]", group))
}

func logTreeItem(isLast bool, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	prefix := "  ├─ "
	if isLast {
		prefix = "  └─ "
	}
	fmt.Fprintln(color.Output, colorMuted.Sprint(prefix)+msg)
	broadcastWebLog("tree", prefix+msg)
}
