package httphandlers

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"share/filehandlers"
	"share/utils"
)

func ContentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	offsetStr := r.URL.Query().Get("offset")
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil || offset < 0 {
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}

	if err := filehandlers.SeekFile(offset); err != nil {
		http.Error(w, "Error seeking file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fileSize := strconv.FormatInt(filehandlers.GetFileSize(), 10)
	w.Header().Set("X-Content-Length", fileSize)

	bufWriter := bufio.NewWriter(w)
	defer bufWriter.Flush()

	buffer := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return // Client disconnected or request cancelled
		default:
			n, err := filehandlers.ReadChunk(ctx, buffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading file: %v", err)
				}
				return
			}
			if n == 0 {
				return
			}
			if _, err := bufWriter.Write(buffer[:n]); err != nil {
				log.Printf("Error writing to response: %v", err)
				return
			}
		}
	}
}

//go:embed index.html
var content embed.FS

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	// Load the embedded index.html file
	data, err := content.ReadFile("index.html")
	if err != nil {
		log.Println("Error reading embedded file:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	memoryUsage, cpuUsage, internetStatus := utils.GetSystemStats()
	stats := map[string]string{
		"memory":   memoryUsage,
		"cpu":      cpuUsage,
		"internet": internetStatus,
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"memory": "%s", "cpu": "%s", "internet": "%s"}`, stats["memory"], stats["cpu"], stats["internet"])
}
