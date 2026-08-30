/*
The main web server.

Before running (go run main.go / go build main.go && ./main):
    1. grabfiles.py must be run so that the project folders are in the 'scs' directory.
    2. There should be an index.html and a 404.html in the 'scs' directory.

Distributed under the MIT software license. See the accompanying file LICENSE or https://opensource.org/license/mit/.
*/

package main

import (
    "fmt"
    "os"
    "log"
    "net/http"
    "path"
)

// adapted from https://stackoverflow.com/a/62747667
func customHandler(fsPath string) http.Handler {
    fs := http.Dir(fsPath)
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        page404, err := os.ReadFile(fsPath + "/404.html")
        if err != nil {
            http.Error(w, "Internal Server Error", 500)
            return
        }

        cleanURL := path.Clean(r.URL.Path)
        _, err = fs.Open(cleanURL)
        if cleanURL == "/static" || os.IsNotExist(err) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            w.WriteHeader(http.StatusNotFound)
            w.Write(page404)
            return
        }

        http.FileServer(fs).ServeHTTP(w, r)
    })
}

func main() {
    port := ":9876"
    fmt.Println("HTTP server is running on port " + port)

    log.Fatal(http.ListenAndServe(port, customHandler("./scs")))
}
