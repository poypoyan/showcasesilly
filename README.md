# showcasesilly
Showcase my silly projects.

By "silly", I mean the whole thing is in a webpage (no server side, just vanilla HTML, JS, CSS). I just need to have access to those without first downloading my Git repositories.

## Setup
1. Run `grabfiles.py`. On each project directory created in `./scs/`, one HTML file will be renamed as `index.html`.
2. Craft your own `index.html` and `404.html` in `./scs/`. The projects must be referenced in links (href) as `/project-dir/`.
3. Either run, or build and run the web server:
```bash
go run main.go
```
or
```bash
go build main.go
./main
```

Note: port is set to `9876`.

## License
Distributed under the MIT software license. See the accompanying
file LICENSE or https://opensource.org/license/mit/.