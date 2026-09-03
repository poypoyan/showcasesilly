/*
The projects grabber.

./grabproj projects.json template.txt > list.txt

Copies all HTML, CSS, JS files (not recursive) to /scs/project/<project-dir-name>.

Distributed under the MIT software license. See the accompanying file LICENSE or https://opensource.org/license/mit/.
*/

package main

import (
    "bytes"
    "encoding/json/v2"
    "fmt"
    "io"
    "log"
    "os"
    "os/exec"
    "strings"
)

func main() {
    args := os.Args[1:]

    if len(args) != 2 {
        fmt.Fprintf(os.Stderr, "Error: Two arguments are required: path to projects JSON and path to template for stdout.\n")
        return
    }

    projJSON, err := os.ReadFile(args[0])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: projects json is not found or inaccessible.\n")
        return
    }

    template, err := os.ReadFile(args[1])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: template is not found or inaccessible.\n")
        return
    }

    process(projJSON, string(template))
}

type Project struct {
    Name string     // name of project, required
    Repo string     // Git repo link, required
    Desc string     // description of project
    Exweb string    // external website
    Exext []string  // additional file extensions to copy
    Rloc string     // location of file/directory inside repo
    Selidx string   // HTML file to select in case there are multiple HTML files
    Rename string   // rename project directory
}

type ProcTempl struct {
    Name string
    Repo string
    Desc string
    Page string
}

func process(p []byte, t string) {
    tempPath := "./.temp-grabproj"
    destPath := "./scs/project"

    // clean before starting
    os.RemoveAll(tempPath)
    os.RemoveAll(destPath)

    // make dirs
    os.Mkdir(tempPath, os.ModePerm)
    os.Mkdir(destPath, os.ModePerm)

    // parse JSON
    var projs []Project
    err := json.Unmarshal(p, &projs)
    if err != nil {
        log.Printf("JSON parsing error: %v\n", err)
        return
    }

    for i, proj := range projs {
        if len(proj.Name) == 0 {
            log.Printf("Project #%d error: Name is empty.\n", i)
            break
        }
        if len(proj.Repo) == 0 {
            log.Printf("Project #%d error: Repo is empty.\n", i)
            break
        }

        var proct ProcTempl
        proct.Name = proj.Name
        proct.Repo = proj.Repo
        proct.Desc = proj.Desc

        // case 1: project is an external website
        if len(proj.Exweb) > 0 {
            proct.Page = proj.Exweb
            processTempl(proct, t)
            continue
        }

        // extract repo name and determine project-dir-name
        repoName := proj.Repo[strings.LastIndex(proj.Repo, "/") + 1:]
        var projDirName string
        if len(proj.Rename) > 0 {
            projDirName = proj.Rename
        } else {
            projDirName = repoName
        }

        // git clone
        tempRepoDir := tempPath + "/" + repoName
        _, err = os.Stat(tempRepoDir)
        if os.IsNotExist(err) {
            os.Mkdir(tempRepoDir, os.ModePerm)
            runGitClone(proj.Repo, tempRepoDir)
        }

        // setup repo and copy directories, and replacement for {Page} in template
        tempProjPath := tempRepoDir + "/" + proj.Rloc
        copyProjDir := destPath + "/" + projDirName
        proct.Page = "/project/" + projDirName + "/"
        os.Mkdir(copyProjDir, os.ModePerm)

        // case 2: project is a single HTML file
        if isHTMLFilePath(proj.Rloc) {
            err = copyFile(tempProjPath, copyProjDir + "/index.html")
            if err != nil {
                log.Printf("Project #%d error: Sole HTML copy error: %v\n", i, err)
                os.RemoveAll(copyProjDir)
                continue
            }
            processTempl(proct, t)
            continue
        }

        // case 3: project is a directory
        listFiles, err := getFiles(tempProjPath, proj.Exext)
        if err != nil {
            log.Printf("Project #%d error: Failed to read directory: %v\n", i, err)
            os.RemoveAll(copyProjDir)
            continue
        }
        idxIdx := getIndexHTML(listFiles, proj.Selidx)
        if idxIdx == -1 {
            log.Printf("Project #%d error: No HTML file found.\n", i)
            os.RemoveAll(copyProjDir)
            continue
        }

        isCopyFail := false
        for j, file := range listFiles {
            if j == idxIdx {
                err = copyFile(tempProjPath + "/" + file, copyProjDir + "/index.html")
            } else {
                err = copyFile(tempProjPath + "/" + file, copyProjDir + "/" + file)
            }
            if err != nil {
                isCopyFail = true
                os.RemoveAll(copyProjDir)
                log.Printf("Project #%d error: Copy error among multiple files: %v.\n", i, err)
                break
            }
        }
        if isCopyFail {
            continue
        }

        processTempl(proct, t)
    }

    // clean temp
    os.RemoveAll(tempPath)
}

func runGitClone(url string, repoDir string) {
    // assumptions: 1) the repo exists and 2) it will always clone the default branch like "main"
    // there is no error handling for this, so if unexpected happens (e.g., asks for username),
    // just Ctrl+C, edit the JSON, and run again.
    cmd := exec.Command("git", "clone", url, repoDir)
    var stderrBuf bytes.Buffer
    cmd.Stderr = &stderrBuf

    err := cmd.Run()
    if err != nil {
        log.Printf("Git, %s\n", stderrBuf.String())
    }
}

func isHTMLFilePath(filePath string) bool {
    return strings.HasSuffix(filePath, ".html") || strings.HasSuffix(filePath, ".htm")
}

func getFiles(p string, exext []string) ([]string, error) {
    allExts := append([]string{"html", "htm", "css", "js"}, exext...)
    var listFiles []string

    entries, err := os.ReadDir(p)
    if err != nil {
        return listFiles, err
    }

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        for _, ext := range allExts {
            if strings.HasSuffix(entry.Name(), "." + ext) {
                listFiles = append(listFiles, entry.Name())
                break
            }
        }
    }
    return listFiles, nil
}

func getIndexHTML(files []string, altFile string) int {
    for i, file := range files {
        if file == altFile || isHTMLFilePath(file) {
            return i
        }
    }
    return -1
}

func copyFile(src string, dst string) error {
    // adapted from https://zetcode.com/golang/copyfile/
    fin, err := os.Open(src)
    if err != nil {
        return err
    }
    defer fin.Close()

    fout, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer fout.Close()

    _, err = io.Copy(fout, fin)
    if err != nil {
        return err
    }
    return nil
}

func processTempl(p ProcTempl, t string) {
    repl := strings.NewReplacer("{Name}", p.Name, "{Repo}", p.Repo, "{Desc}", p.Desc, "{Page}", p.Page)
    fmt.Print(repl.Replace(t))
}
