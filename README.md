# Go-Readability [![Go Reference][go-ref-badge]][go-ref]

Go-Readability is a Go package that find the main readable content and the metadata from a HTML page. It works by removing clutter like buttons, ads, background images, script, etc.

This is a fork of [github.com/go-shiori/go-readability](https://github.com/go-shiori/go-readability) originally written by Radhi Fadlillah and maintained by Felipe Martin and GitHub contributors. For more information about the changes in this fork, see [FORK.md](./FORK.md).

Radhi Fadlillah initially ported [Readability.js] line-by-line to Go to make sure it looks and works as similar as possible. This way, hopefully all web page that can be parsed by Readability.js are parse-able by go-readability as well.

This module is compatible with Readability.js v0.6.0.

## Installation

**Note:** you are viewing documentation for version 0, which is API-compatible with `github.com/go-shiori/go-readability`. The development of this project continues in the [v2 branch], which you should choose for _best speed and memory efficiency_, with API-breaking changes being that some `Article` fields were converted to methods.

To add this package to your project, use `go get`:

```
go get -u codeberg.org/readeck/go-readability
```

And to get the [v2 branch] instead:

```
go get -u codeberg.org/readeck/go-readability/v2
```

## Example

```go
package main

import (
	"fmt"
	"log"
	"os"

	readability "codeberg.org/readeck/go-readability"
)

func main() {
	srcFile, err := os.Open("index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	baseURL, _ := url.Parse("https://example.com/path/to/article")
	article, err := readability.FromReader(srcFile, baseURL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found article with title %q\n\n", article.Title)
	// Print the parsed, cleaned-up HTML markup of the article.
	fmt.Println(article.Content)
}
```

## Command Line Usage

You can also use `go-readability` as command-line tool:

```
go install codeberg.org/readeck/go-readability/cmd/go-readability@latest
```

Now you can use it by running `go-readability` in your terminal :

```
$ go-readability -h

go-readability is a parser that extracts article contents from a web page.
The source can be a URL or a filesystem path to a HTML file.
Pass "-" or no argument to read the HTML document from standard input.
Use "--http :0" to automatically choose an available port for the HTTP server.

Usage:
  go-readability [<flags>...] [<url> | <file> | -]

Flags:
  -h, --help          help for go-readability
  -l, --http string   start the http server at the specified address
  -m, --metadata      only print the page's metadata
  -t, --text          only print the page's text
```


[go-ref]: https://pkg.go.dev/codeberg.org/readeck/go-readability
[go-ref-badge]: https://img.shields.io/static/v1?label=&message=Reference&color=007d9c&logo=go&logoColor=white
[readability.js]: https://github.com/mozilla/readability
[v2 branch]: https://codeberg.org/readeck/go-readability/src/branch/v2
