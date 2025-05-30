package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	addr          string
	storageFolder string
	token         string
)

const uploadFormHTML = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Upload</title></head>
<body>
  <form enctype="multipart/form-data" action="/upload" method="post">
    <input type="file" name="file" />
    <input type="submit" value="Upload" />
  </form>
</body>
</html>`

// --- Middleware ---
func withTokenValidation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 只有当token不为空时才做验证
		if token != "" && r.Header.Get("token") != token {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// --- Handlers ---
func handleUploadForm(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	w.Write([]byte(uploadFormHTML))
}

func handleUploadFile(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	today := time.Now().Format("2006-01-02")
	uploadDir := path.Join(storageFolder, today)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
		return
	}

	dstPath := path.Join(uploadDir, handler.Filename)

	dstFile, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Error creating destination file", http.StatusInternalServerError)
		return
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	log.Printf("[UPLOAD] File saved to: %s\n", dstPath)
	fmt.Fprintf(w, "File uploaded to: %s\n", dstPath)
}

func handleLogRequest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	clientIP := strings.Split(r.RemoteAddr, ":")[0]
	filename := fmt.Sprintf("%s_%s.txt", clientIP, time.Now().Format("20060102_150405"))
	filePath := path.Join(storageFolder, filename)

	f, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving request log", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	r.Write(f)
	log.Printf("[LOG] Full request saved: %s\n", filename)
	w.Write([]byte(fmt.Sprintf("Request saved as: %s\n", filename)))
}

func handleFileServer(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	http.FileServer(http.Dir(storageFolder)).ServeHTTP(w, r)
}

// --- Routing ---
func setupRoutes() {
	http.HandleFunc("/upload-form", withTokenValidation(handleUploadForm))
	http.HandleFunc("/upload", withTokenValidation(handleUploadFile))
	http.HandleFunc("/log-request", withTokenValidation(handleLogRequest))
	http.Handle("/files/", http.StripPrefix("/files", withTokenValidation(handleFileServer)))
}

// --- CLI 参数解析 ---
func parseFlags() {
	flag.StringVar(&addr, "listen", ":58080", "Listening address (e.g., :58080)")
	flag.StringVar(&storageFolder, "storage", "./data", "Folder to store uploaded files")
	flag.StringVar(&token, "token", "", "Access token (optional; if empty, no token check)")
	flag.Parse()
}

// --- 实用工具函数 ---
func logRequest(r *http.Request) {
	log.Printf("[REQ] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
}

// --- 主程序入口 ---
func main() {
	parseFlags()

	if _, err := os.Stat(storageFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(storageFolder, 0755); err != nil {
			log.Fatalf("Failed to create storage folder: %v", err)
		}
	}

	setupRoutes()

	log.Printf("Server started at %s", addr)
	log.Printf("Upload directory: %s", storageFolder)
	if token != "" {
		log.Printf("Token is enabled")
	} else {
		log.Printf("Token is disabled (no validation)")
	}
	log.Fatal(http.ListenAndServe(addr, nil))
}
