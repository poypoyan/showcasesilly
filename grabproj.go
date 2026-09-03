/*
The projects grabber.

./grabproj projects.json template.txt > list.txt

Copies all HTML, CSS, JS files (not recursive) to /scs/projects/<proj-dir>.

Distributed under the MIT software license. See the accompanying file LICENSE or https://opensource.org/license/mit/.
*/

package main

import (
    "bytes"
    "encoding/json/v2"
    "fmt"
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
    RLoc string     // location of file/directory inside repo
    Selidx string   // HTML file to select is case there are multiple HTML files
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
        log.Println("JSON parsing error: %s", err.Error())
        return
    }

    for i, proj := range projs {
        if len(proj.Name) == 0 {
            log.Println("Project #%d error: Name is empty.", i)
            break
        }
        if len(proj.Repo) == 0 {
            log.Println("Project #%d error: Repo is empty.", i)
            break
        }

        var proct ProcTempl
        proct.Name = proj.Name
        proct.Repo = proj.Repo
        proct.Desc = proj.Desc
        proct.Page = "/404"   // for identifying if there's nothing to showcase

        // case 1: project is an external website
        if len(proj.Exweb) > 0 {
            proct.Page = proj.Exweb
            processTempl(proct, t)
            continue
        }

        // extract repo name
        repoName := proj.Repo[strings.LastIndex(proj.Repo, "/") + 1:]
        var projName string
        if len(proj.Rename) > 0 {
            projName = proj.Rename
        } else {
            projName = repoName
        }
        proct.Page = "/" + projName

        // git clone
        tempRepoDir := tempPath + "/" + repoName
        _, err = os.Stat(tempRepoDir)
        if os.IsNotExist(err) {
            os.Mkdir(tempRepoDir, os.ModePerm)
            runGitClone(proj.Repo, tempRepoDir)
        }

        // TODO: copy files

        processTempl(proct, t)
    }

    // clean temp
    // os.RemoveAll(tempPath)
}

func processTempl(p ProcTempl, t string) {
    repl := strings.NewReplacer("{Name}", p.Name, "{Repo}", p.Repo, "{Desc}", p.Desc, "{Page}", p.Page)
    fmt.Print(repl.Replace(t))
}

func runGitClone(url string, repoDir string) {
    // assumptions: 1) the repo exists and 2) it will always clone the main branch
    // there is no error handling for this, so if unexpected happens (e.g., asks for username),
    // just Ctrl+C, edit the JSON, and run again.
    cmd := exec.Command("git", "clone", url, repoDir)
    var stderrBuf bytes.Buffer
    cmd.Stderr = &stderrBuf

    err := cmd.Run()
    if err != nil {
        log.Println("Git, " + stderrBuf.String())
    }
}

