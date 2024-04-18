package main

import (
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize(func() {
		yalog.SetLevelInfo()
		yalog.DisableJSONLogger()
	})
}

var root = &cobra.Command{
	Use:   "lindows",
	Short: "Lindows is a WebRTC-based remote desktop service",
	Long:  "Lindows is a WebRTC-based remote desktop service",
}

func Execute() {
	err := root.Execute()
	if err != nil {
		yalog.Fatal("Failed to execute command", "error", err)
	}
}
