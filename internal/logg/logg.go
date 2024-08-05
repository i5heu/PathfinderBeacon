package logg

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
)

func InitLogger() *zap.Logger {

	absPath := "/logs/server.log"
	err := os.MkdirAll("/logs", 0755)
	if err != nil {
		fmt.Printf("Failed to create logs folder: %v \n", err)
		return nil
	}

	cfg := zap.NewProductionConfig()

	fmt.Println(absPath)

	cfg.OutputPaths = []string{
		absPath,
	}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Failed to create logger: %s\n", err)
	}

	logger.Info("Logger initialized")

	return logger
}
