// main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/mdp/qrterminal/v3"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"share/utils"
	"syscall"
	"time"

	"share/filehandlers"
	"share/httphandlers"
)

var filePath string

func main() {
	parseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize file
	err := filehandlers.InitializeFile(filePath)
	if err != nil {
		log.Fatalf("Error initializing file: %v", err)
	}
	defer filehandlers.CloseFile()

	localIP, randomPort := utils.GetLocalIPAddressAndPort()

	// Set up HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", randomPort),
		Handler: setupHandlers(ctx),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	fileURL := fmt.Sprintf("http://%s:%d", localIP, randomPort)
	viewUrl := fileURL + "/view"

	log.Printf("Server started on %s", viewUrl)

	fmt.Println("Scan the QR code below to access the file:")
	qrterminal.GenerateWithConfig(viewUrl, qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
	})

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func parseFlags() {
	flag.StringVar(&filePath, "file", "", "Path to the file to be served")
	flag.Parse()
}

func setupHandlers(ctx context.Context) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/content", func(w http.ResponseWriter, r *http.Request) {
		httphandlers.ContentHandler(ctx, w, r)
	})
	mux.HandleFunc("/view", httphandlers.ViewHandler)
	mux.HandleFunc("/stats", httphandlers.StatsHandler)
	return mux
}
