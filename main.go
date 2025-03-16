package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/NesterovYehor/Crawler/internal/crawler"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "URL to crawl",
	Short: "Website URL we want to crawl",
	Run: func(cmd *cobra.Command, args []string) {
		maxConcurrency := 3

		if len(args) > 1 {
			var err error
			maxConcurrency, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println("Invalid concurrency value:", args[1])
				os.Exit(1)
				return
			}
		}

		pages := map[string]int{}

		pool, err := crawler.NewWorkerPool("urls.txt", maxConcurrency, pages)
		pool.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		return
	}
}
