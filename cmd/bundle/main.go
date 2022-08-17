package main

import (
	"flag"
	"fmt"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/samber/lo"
)
import "os"

func main() {
	watch := flag.Bool("watch", false, "watch files")
	production := flag.Bool("production", false, "production")
	flag.Parse()

	result := api.Build(api.BuildOptions{
		Bundle:            true,
		EntryPoints:       []string{"www/js/index.tsx"},
		JSXMode:           api.JSXModeTransform,
		LogLevel:          api.LogLevelInfo,
		MinifyIdentifiers: *production,
		MinifySyntax:      *production,
		MinifyWhitespace:  *production,
		Outdir:            "static",
		Sourcemap:         api.SourceMapLinked,
		Watch:             lo.If(*watch, &api.WatchMode{}).Else(nil),
		Write:             true,
	})

	if len(result.Errors) > 0 {
		fmt.Println("errors", result.Errors)
		os.Exit(1)
	}

	if *watch {
		<-make(chan struct{})
	}
}
