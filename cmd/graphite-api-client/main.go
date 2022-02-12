package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	graphiteUsername string
	graphitePassword string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   os.Args[0],
		Short: "graphite-api-client is a client for Graphite API",
	}

	graphiteUsername = os.Getenv("GRAPHITE_USERNAME")
	graphitePassword = os.Getenv("GRAPHITE_PASSWORD")

	renderCmd(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}
