package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kr/pretty"
	graphiteapi "github.com/msaf1980/graphite-api-client"
	"github.com/spf13/cobra"
)

type RootCfg struct {
	Base          string
	From          string
	Until         string
	Targets       StringSlice
	MaxDataPoints int
}

var rootCfg = RootCfg{}

func renderRun(*cobra.Command, []string) {
	if len(rootCfg.Base) == 0 {
		log.Fatalf("base address not set")
	}
	if len(rootCfg.Targets) == 0 {
		log.Fatalf("targets not set")
	}
	if rootCfg.MaxDataPoints < 0 {
		log.Fatalf("max data points must be >= 0")
	}

	q := graphiteapi.NewRenderQuery(rootCfg.Base, rootCfg.From, rootCfg.Until, rootCfg.Targets, rootCfg.MaxDataPoints)
	if graphiteUsername != "" {
		q.SetBasicAuth(graphiteUsername, graphitePassword)
	}

	result, err := q.Request(context.Background())
	if err != nil {
		log.Fatalf("Render query error: %s", err)
	}

	fmt.Printf("%# v\n", pretty.Formatter(result))
}

func renderCmd(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Get graphite api render function",
		Run:   renderRun,
	}

	cmd.Flags().StringVarP(&rootCfg.Base, "base", "b", "http://127.0.0.1:8888", "base address (for basic auth set GRAPHITE_USERNAME and GRAPHITE_PASSWORD env vars)")
	cmd.Flags().VarP(&rootCfg.Targets, "targets", "t", "targets")
	cmd.Flags().StringVarP(&rootCfg.From, "from", "f", "", "from")
	cmd.Flags().StringVarP(&rootCfg.Until, "until", "u", "", "until")
	cmd.Flags().IntVar(&rootCfg.MaxDataPoints, "m", 0, "max data points")

	rootCmd.AddCommand(cmd)
}
