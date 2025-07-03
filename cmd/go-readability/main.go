package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	nurl "net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	readability "codeberg.org/readeck/go-readability"
	"github.com/go-shiori/dom"
	flag "github.com/spf13/pflag"
)

// The User-Agent string used when fetching remote URLs.
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"

const index = `<!DOCTYPE HTML>
<html>
 <head>
  <meta charset="utf-8">
  <title>go-readability</title>
 </head>
 <body>
 <form action="/" style="width:80%">
  <fieldset>
   <legend>Get readability content</legend>
   <p><label for="url">URL </label><input type="url" name="url" style="width:90%"></p>
   <p><input type="checkbox" name="text" value="true">text only</p>
   <p><input type="checkbox" name="metadata" value="true">only get the page's metadata</p>
  </fieldset>
  <p><input type="submit"></p>
 </form>
 </body>
</html>`

var (
	httpListen   string
	metadataOnly bool
	textOnly     bool
	verbose      bool
)

func printUsage(w io.Writer, flags *flag.FlagSet) {
	fmt.Fprintln(w, "Usage:\n  go-readability [flags...] {<url> | <file>}\n\nFlags:")
	fmt.Fprintln(w, flags.FlagUsages())
}

func main() {
	flags := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	// Override pflag's builtin Usage implementation which unconditionally prints to stderr.
	flags.Usage = func() {}

	flags.StringVarP(&httpListen, "http", "l", "", "start the http server at the specified address")
	flags.BoolVarP(&metadataOnly, "metadata", "m", false, "only print the page's metadata")
	flags.BoolVarP(&textOnly, "text", "t", false, "only print the page's text")
	flags.BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	if err := flags.Parse(os.Args[1:]); err != nil || flag.NArg() < 1 {
		if errors.Is(err, flag.ErrHelp) {
			// When explicitly asked for command help, print usage string to stdout.
			fmt.Fprintln(os.Stdout,
				"go-readability is a parser that extracts article contents from a web page.\n"+
					"The source can be a URL or a filesystem path to a HTML file.")
			fmt.Fprintln(os.Stdout)
			printUsage(os.Stdout, flags)
			return
		} else if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		printUsage(os.Stderr, flags)
		os.Exit(2)
	}

	err := rootCmdHandler(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmdHandler(source string) error {
	if httpListen != "" {
		// Start HTTP server
		http.HandleFunc("/", httpHandler)
		log.Println("Starting HTTP server at", httpListen)
		return http.ListenAndServe(httpListen, nil)
	}

	content, err := getContent(source, metadataOnly, textOnly, verbose)
	if err != nil {
		return err
	}

	_, err = fmt.Println(content)
	return err
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	metadataOnly, _ := strconv.ParseBool(r.URL.Query().Get("metadata"))
	textOnly, _ := strconv.ParseBool(r.URL.Query().Get("text"))
	url := r.URL.Query().Get("url")
	if url == "" {
		if _, err := w.Write([]byte(index)); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		log.Println("process URL", url)
		content, err := getContent(url, metadataOnly, textOnly, false)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if metadataOnly {
			w.Header().Set("Content-Type", "application/json")
		} else if textOnly {
			w.Header().Set("Content-Type", "text/plain")
		}
		if _, err := w.Write([]byte(content)); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func getContent(srcPath string, metadataOnly, textOnly, verbose bool) (string, error) {
	// Open or fetch web page that will be parsed
	var (
		pageURL   *nurl.URL
		srcReader io.Reader
	)

	if _, isURL := validateURL(srcPath); isURL {
		req, err := http.NewRequest("GET", srcPath, nil)
		if err != nil {
			return "", fmt.Errorf("failed to construct the request: %v", err)
		}
		req.Header.Add("User-Agent", userAgent)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to fetch web page: %v", err)
		}
		defer resp.Body.Close()

		pageURL = resp.Request.URL
		srcReader = resp.Body
	} else {
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return "", fmt.Errorf("failed to open source file: %v", err)
		}
		defer srcFile.Close()

		pageURL, _ = nurl.ParseRequestURI("http://fakehost.com")
		srcReader = srcFile
	}

	doc, err := dom.Parse(srcReader)
	if err != nil {
		return "", fmt.Errorf("HTML parse error: %w", err)
	}

	// Make sure the page is readable
	if !readability.CheckDocument(doc) {
		return "", fmt.Errorf("failed to parse page: the page is not readable")
	}

	parser := readability.NewParser()
	parser.Debug = verbose

	// Get readable content from the reader
	article, err := parser.ParseAndMutate(doc, pageURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse page: %v", err)
	}

	// Return the article (or its metadata)
	if metadataOnly {
		metadata := map[string]interface{}{
			"title":   article.Title,
			"byline":  article.Byline,
			"excerpt": article.Excerpt,
			"image":   article.Image,
			"favicon": article.Favicon,
		}

		prettyJSON, err := json.MarshalIndent(&metadata, "", "    ")
		if err != nil {
			return "", fmt.Errorf("failed to write metadata file: %v", err)
		}

		return string(prettyJSON), nil
	}

	if textOnly {
		return article.TextContent, nil
	}

	return article.Content, nil
}

func validateURL(path string) (*nurl.URL, bool) {
	url, err := nurl.ParseRequestURI(path)
	return url, err == nil && strings.HasPrefix(url.Scheme, "http")
}
