package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var storageFolder string
var addr string
var token string
var form = `<!DOCTYPE html>
<html lang="en">
   <head>
      <meta charset="UTF-8" />
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta http-equiv="X-UA-Compatible" content="ie=edge" />
   </head>
   <body>
      <form enctype="multipart/form-data" action="/p" method="post">
         <input type="file" name="file" />
         <input type="submit" value="Upload" />
      </form>
   </body>
</html>`

func isValidToken(req *http.Request) bool {
	reqToken := req.Header.Get("token")
	return reqToken == token
}

func isValidTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidToken(r) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func exfilGet(w http.ResponseWriter, req *http.Request) {
	if !isValidToken(req) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	host := strings.Split(req.RemoteAddr, ":")
	filename := fmt.Sprintf("%s_%s.txt", host[0], time.Now().Format("2006-01-02_15-04-05"))
	filePath := path.Join(storageFolder, filename)

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	req.Write(out)
	fmt.Printf("[*] Request Stored (%s)\n", filename)

	// Return filename and other information in the response
	response := fmt.Sprintf("File uploaded successfully. Filename: %s", filename)
	w.Write([]byte(response))
}

func uploadForm(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, form)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	ioutil.WriteFile(path.Join(storageFolder, handler.Filename), fileBytes, 0644)

	fmt.Fprintf(w, "Done\n")
	fmt.Printf("[*] File Uploaded (%s)\n", handler.Filename)
}

func setupRoutes() {
	fileHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isValidToken(r) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		http.FileServer(http.Dir(storageFolder)).ServeHTTP(w, r)
	})

	http.HandleFunc("/", isValidTokenMiddleware(uploadForm))
	http.HandleFunc("/p", isValidTokenMiddleware(uploadFile))
	http.HandleFunc("/g", isValidTokenMiddleware(exfilGet))

	http.Handle("/l/", http.StripPrefix("/l", isValidTokenMiddleware(fileHandler)))

	if _, err := os.Stat("httpuploadGO.csr"); err == nil {
		log.Fatal(http.ListenAndServeTLS(addr, "httpuploadGO.csr", "httpuploadGO.key", nil))
	} else {
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}

func parseFlags() {
	flag.StringVar(&addr, "port", "58080", "Specify the listening port")
	flag.StringVar(&storageFolder, "path", ".", "Specify the storage path")
	flag.StringVar(&token, "token", "", "Specify the header token value")
	flag.Parse()

	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}
}

func main() {
	parseFlags()

	fmt.Printf("[+] Server Running...\n")
	fmt.Printf("[+] Settings: Addr '%s'; Folder '%s'; Token '%s'\n", addr, storageFolder, token)
	fmt.Printf("[+] Instructions: '/' directory quick upload, '/p' directory to manually build and upload files, '/l' directory gets the current folder contents")

	if _, err := os.Stat(storageFolder); os.IsNotExist(err) {
		err := os.Mkdir(storageFolder, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
	//fmt.Println("Before setupRoutes")
	setupRoutes()
}
