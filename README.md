# showcasesilly
Showcase my silly projects.

By "silly", I mean the whole thing is in a webpage (no server side, just vanilla HTML, JS, CSS). I just need to have access to those without first downloading my Git repositories.

## Setup
1. Build the project grabber (`grabproj.go`) and the web server (`main.go`):
```bash
go build grabproj.go
go build main.go
```
2. Run the project grabber (`./grabproj`). On each project directory created in `./scs/`, one HTML file will be renamed as `index.html`. Also, you can input a template file to "display" the project in stdout; this is useful for creating HTML list.
3. Craft your own `index.html` and `404.html` in `./scs/`. The projects must be referenced in links (href) as `/project-dir/`.
4. Run the web server (`./main`).

Note: port is set to `9876`.

## License
Distributed under the MIT software license. See the accompanying
file LICENSE or https://opensource.org/license/mit/.