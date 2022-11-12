package main

import (
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

func uptime() time.Duration {
	return time.Since(startTime)
}

func properExit() {
	// Not formatting string because I only want the exit message to be red.
	log.Println(color.HiRedString("[EXIT IN 15 SECONDS]"), " Uptime was", durafmt.Parse(time.Since(startTime)).String(), "...")
	log.Println(color.HiCyanString("--------------------------------------------------------------------------------"))
	time.Sleep(15 * time.Second)
	os.Exit(1)
}
