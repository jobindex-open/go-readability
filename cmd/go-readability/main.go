package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	nurl "net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-shiori/dom"
	readability "github.com/jobindex-open/go-readability"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
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

const stdinPath = "-"

var (
	httpListen   string
	metadataOnly bool
	textOnly     bool
	verbose      bool
	force        bool
)

func printUsage(w io.Writer, flags *flag.FlagSet) {
	_, _ = fmt.Fprintln(w, "Usage:\n  go-readability [<flags>...] [<url> | <file> | -]\n\nFlags:")
	_, _ = fmt.Fprintln(w, flags.FlagUsages())
}

type statusErr int

func (s statusErr) Error() string {
	return fmt.Sprintf("exit status %d", s)
}

func mainRun(args []string) error {
	flags := flag.NewFlagSet(filepath.Base(args[0]), flag.ContinueOnError)
	// Override pflag's builtin Usage implementation which unconditionally prints to stderr.
	flags.Usage = func() {}

	flags.StringVarP(&httpListen, "http", "l", "", "start the http server at the specified address (example: \":3000\")")
	flags.BoolVarP(&metadataOnly, "metadata", "m", false, "only print the page's metadata")
	flags.BoolVarP(&textOnly, "text", "t", false, "only print the page's text")
	flags.BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	flags.BoolVarP(&force, "force", "f", false, "continue parsing documents that failed the readerable check")

	if err := flags.Parse(args[1:]); err != nil || flags.NArg() > 1 {
		if errors.Is(err, flag.ErrHelp) {
			// When explicitly asked for command help, print usage string to stdout.
			_, _ = fmt.Fprintln(os.Stdout,
				"go-readability is a parser that extracts article contents from a web page.\n"+
					"The source can be a URL or a filesystem path to a HTML file.\n"+
					"Pass \"-\" or no argument to read the HTML document from standard input.\n"+
					"Use \"--http :0\" to automatically choose an available port for the HTTP server.")
			_, _ = fmt.Fprintln(os.Stdout)
			printUsage(os.Stdout, flags)
			return nil
		} else if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		printUsage(os.Stderr, flags)
		return statusErr(2)
	}

	srcPath := stdinPath
	if flags.NArg() > 0 {
		srcPath = flags.Arg(0)
	}

	err := rootCmdHandler(srcPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return statusErr(1)
	}
	return nil
}

func main() {
	err := mainRun(os.Args)
	if err != nil {
		exitStatus := 1
		if s, ok := err.(statusErr); ok {
			exitStatus = int(s)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(exitStatus)
	}
}

func rootCmdHandler(srcPath string) error {
	if httpListen != "" {
		nl, err := net.Listen("tcp", httpListen)
		if err != nil {
			return err
		}
		localhost := "localhost"
		addr := nl.Addr().(*net.TCPAddr)
		if addr.IP.String() != "::" {
			localhost = addr.IP.String()
		}
		fmt.Fprintf(os.Stderr, "Starting HTTP server at http://%s:%d\n", localhost, addr.Port)
		http.HandleFunc("/", httpHandler)
		server := http.Server{Handler: http.DefaultServeMux}
		return server.Serve(nl)
	}

	content, err := getContent(srcPath, metadataOnly, textOnly)
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
		content, err := getContent(url, metadataOnly, textOnly)
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

func getContent(srcPath string, metadataOnly, textOnly bool) (string, error) {
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
		if srcPath == stdinPath {
			defer os.Stdin.Close()
			srcReader = os.Stdin
		} else {
			srcFile, err := os.Open(srcPath)
			if err != nil {
				return "", fmt.Errorf("failed to open source file: %v", err)
			}
			defer srcFile.Close()
			srcReader = srcFile
		}
		pageURL, _ = nurl.ParseRequestURI("http://fakehost.com")
	}

	doc, err := dom.Parse(srcReader)
	if err != nil {
		return "", fmt.Errorf("HTML parse error: %w", err)
	}

	// Make sure the page is "readerable"
	if !readability.CheckDocument(doc) {
		if force {
			fmt.Fprintf(os.Stderr, "warning: the page might not have article contents\n")
		} else {
			return "", errors.New("failed to detect readable content on the page")
		}
	}

	parser := readability.NewParser()
	if verbose {
		parser.Logger = newLogger(os.Stderr)
	}

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

func newLogger(w *os.File) *slog.Logger {
	return slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: "15:04:05.000",
			NoColor:    !isatty.IsTerminal(w.Fd()),
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Value.Kind() == slog.KindAny {
					if _, ok := a.Value.Any().(error); ok {
						return tint.Attr(9, a)
					}
				}
				if a.Value.Kind() == slog.KindFloat64 {
					return slog.String(a.Key, fmt.Sprintf("%05.1f", a.Value.Float64()))
				}
				return a
			},
		}),
	)
}
