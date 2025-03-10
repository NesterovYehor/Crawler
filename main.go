package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/crawler"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "URL to crawl",
	Short: "Website URL we want to crawl",
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure the number of arguments is correct
		if len(args) < 1 || len(args) > 3 {
			fmt.Println("Usage: go run . <URL> <maxConcurrencies> <maxPages>")
			os.Exit(1)
			return
		}

		// Parse the URL (first argument)
		rawURL := args[0]

		// Default values for concurrency and pages
		maxConcurrency := 3
		maxPages := 25

		// Parse the second argument for maxConcurrency if provided
		if len(args) > 1 {
			var err error
			maxConcurrency, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println("Invalid concurrency value:", args[1])
				os.Exit(1)
				return
			}
		}

		// Parse the third argument for maxPages if provided
		if len(args) > 2 {
			var err error
			maxPages, err = strconv.Atoi(args[2])
			if err != nil {
				fmt.Println("Invalid pages value:", args[2])
				os.Exit(1)
				return
			}
		}

		// Create a new config with the parsed values
		cfg, err := config.NewConfig(rawURL, maxConcurrency, maxPages)
		if err != nil {
			fmt.Println(err)
			return
		}

		cfg.Wg.Add(1)
		err = crawler.CrawlPage(rawURL, cfg)
		if err != nil {
			fmt.Println(err)
			return
		}
		cfg.Wg.Wait()
		fmt.Printf(`
=============================
  REPORT for %v
=============================
        `, rawURL)
		keys := make([]string, 0, len(cfg.Pages))

		for key := range cfg.Pages {
			keys = append(keys, key)
		}
		sort.SliceStable(keys, func(i, j int) bool {
			return cfg.Pages[keys[i]] < cfg.Pages[keys[j]]
		})

		sort.Strings(keys)
		for _, key := range keys {
			fmt.Printf("Found %v internal links to %v\n", cfg.Pages[key], key)
		}
	},
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		return
	}
}
