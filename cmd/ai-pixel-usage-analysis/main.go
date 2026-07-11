package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ai-project-hub-sg/ai-pixel-usage-analysis/internal/app"
)

func main() {
	configPath := flag.String("config", "config.toml", "configuration file")
	envPath := flag.String("env", ".env", "dotenv file")
	databasePath := flag.String("database", "data/analysis.db", "SQLite database")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx, app.Options{ConfigPath: *configPath, EnvPath: *envPath, DatabasePath: *databasePath}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
