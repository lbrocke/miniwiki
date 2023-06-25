package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	wikilink "github.com/abhinav/goldmark-wikilink"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"golang.org/x/crypto/bcrypt"
)

type Wiki struct {
	Name     string
	PassHash string
	Editable bool
	Dir      string
}

type WikiPage struct {
	WikiName     string
	WikiEditable bool
	PageName     string
	Body         string
	EditPage     bool
}

type WikiLinkResolver struct{}

const HTTPTextInternalServerError = "500 internal server error"

var validPagePath = regexp.MustCompile("^/([a-zA-Z0-9_-]*)$")
var validEditPath = regexp.MustCompile("^/e/([a-zA-Z0-9_-]+)$")

var templatePage = template.Must(template.New("edit").Funcs(template.FuncMap{
	"toHTML": func(str string) template.HTML {
		// The toHTML template function prints to given string as actual HTML
		// without escaping e.g. < and > to &lt; and &gt;
		return template.HTML(str)
	},
}).Parse(`
<!DOCTYPE html>
<html lang=en>
  <head>
    <title>{{ .PageName }} :: {{ .WikiName }}</title>

    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=1">

    <link rel="icon" href="data:,">

    <style>
      body {
        max-width: 30em;
        margin: 0 auto;
        font-size: 1.2em;
        line-height: 1.5em;
        padding: 1em 0;
        font-family: Times New Roman, Times, Source Serif Pro, serif, Apple Color Emoji, Segoe UI Emoji, Segoe UI Symbol;
        color: #444;
      }
      a {
        color: #0066CC;
      }
      header {
        margin-bottom: 1em;
      }
      header a, a:visited, a:hover, a:active {
        text-decoration: none;
      }
      header a:visited, a:hover, a:active {
        color: #0066CC;
      }
      header div#edit {
        float: right;
      }
      h1, h2, h3, h4, h5, h6 {
        line-height: 1.2;
      }
      pre {
        background: #efefef;
        font-family: Menlo, Consolas, Monaco, Liberation Mono, Lucida Console, monospace;
        padding: .5em;
        margin: 0;
		overflow: auto;
      }
      img {
        max-width: 100%;
      }
      nav {
        background: #ddedff;
        padding: 1em;
        margin-bottom: 1em;
      }
      nav > p {
        margin: 0;
      }
      nav > ul, ol {
        padding-left: 1em;
        margin: 0;
      }
      blockquote {
        border-left: 0.3em solid #efefef;
        margin-left: 0;
        padding-left: 1em;
      }
      table {
        width: 100%;
        border-collapse: collapse;
      }
      table th {
        border-bottom: 1px solid #444;
      }
      table td {
        padding: 0.5em;
      }
      textarea {
        width: 100%;
        -webkit-box-sizing: border-box;
        -moz-box-sizing: border-box;
        box-sizing: border-box;
      }
    </style>
  </head>
  <body>

  <header>
    <b>{{ .PageName }}</b> :: <a href="/">{{ .WikiName }}</a>

    {{ if .WikiEditable }}
      <div id="edit">
      {{ if .EditPage }}
        <input form="form" type="password" autocomplete="current-password" name="pass" placeholder="Password">
        <input form="form" type="submit" value="✓">
      {{ else }}
        <a href="/e/{{ .PageName }}">✎</a>
      {{ end }}
      </div>
    {{ end }}
  </header>

  {{ if .EditPage }}
    <form id="form" action="/e/{{ .PageName }}" method="post">
    <textarea autofocus rows="30" name="body">{{ .Body }}</textarea>
    </form>
  {{ else }}
    {{ .Body | toHTML }}
  {{ end }}
  </body>
</html>
`))

func (WikiLinkResolver) ResolveWikilink(n *wikilink.Node) ([]byte, error) {
	link, err := wikilink.DefaultResolver.ResolveWikilink(n)
	if err != nil {
		return nil, err
	}

	// The default resolver simply returns the destination with a .html
	// ending, remove this file extension.
	return []byte(strings.TrimSuffix(string(link), ".html")), nil
}

func (wiki *Wiki) getPageFilePath(name string) string {
	return filepath.Clean(fmt.Sprintf("%s/%s.md", wiki.Dir, name))
}

func (wiki Wiki) editPage(w http.ResponseWriter, r *http.Request) {
	match := validEditPath.FindStringSubmatch(r.URL.Path)

	if match == nil {
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}

	// page can't be empty string, as it is not allowed by the regex
	page := match[1]

	if !wiki.Editable {
		// If the wiki is not editable (no password was configured),
		// redirect to content page.
		http.Redirect(w, r, fmt.Sprintf("/%s", page), http.StatusFound)
		return
	}

	file := wiki.getPageFilePath(page)

	switch r.Method {
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, HTTPTextInternalServerError, 500)
			return
		}

		body := r.FormValue("body")
		pass := r.FormValue("pass")

		if bcrypt.CompareHashAndPassword([]byte(wiki.PassHash), []byte(pass)) != nil {
			// Incorrect password, show edit form again with submitted content
			// so that nothing gets lost
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			templatePage.Execute(w, WikiPage{
				WikiName:     wiki.Name,
				WikiEditable: wiki.Editable,
				PageName:     page,
				Body:         body,
				EditPage:     true,
			})
			return
		}

		if body == "" {
			// An empty body means that the entire page should be deleted
			if err := os.Remove(file); err != nil {
				http.Error(w, HTTPTextInternalServerError, 500)
				return
			}

			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
			return
		}

		if err := os.WriteFile(file, []byte(body), 0640); err != nil {
			http.Error(w, HTTPTextInternalServerError, 500)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/%s", page), http.StatusFound)
	default: //lint:ignore ST1015 Other methods are ignored and handled as GET
	case http.MethodGet:
		content, err := os.ReadFile(file)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			// It's fine that the file does not exist. The HTML form
			// will show an empty textarea, the wiki page will be created
			// on first save.
			// In case another error occurred however, an error should be
			// returned. Otherwise the contents of the existing but somehow
			// not readable file could be unintentionally overwritten.
			http.Error(w, HTTPTextInternalServerError, 500)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templatePage.Execute(w, WikiPage{
			WikiName:     wiki.Name,
			WikiEditable: wiki.Editable,
			PageName:     page,
			Body:         string(content),
			EditPage:     true,
		})
	}
}

func (wiki Wiki) showPage(w http.ResponseWriter, r *http.Request) {
	match := validPagePath.FindStringSubmatch(r.URL.Path)

	// Invalid path (illegal page name), redirect to home page
	if match == nil {
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}

	page := match[1]
	if page == "" {
		page = "home"
	}

	file := wiki.getPageFilePath(page)
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		http.Redirect(w, r, fmt.Sprintf("/e/%s", page), http.StatusFound)
		return
	}

	content, err := os.ReadFile(file)
	if err != nil {
		http.Error(w, HTTPTextInternalServerError, 500)
		return
	}

	var htmlBuf bytes.Buffer
	if err := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			&wikilink.Extender{
				Resolver: WikiLinkResolver{},
			},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	).Convert([]byte(content), &htmlBuf); err != nil {
		http.Error(w, HTTPTextInternalServerError, 500)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templatePage.Execute(w, WikiPage{
		WikiName:     wiki.Name,
		WikiEditable: wiki.Editable,
		PageName:     page,
		Body:         htmlBuf.String(),
		EditPage:     false,
	})
}

func main() {
	const DefaultAddr = ":8080"
	const DefaultName = "wiki"
	const DefaultPass = ""
	const DefaultDir = "./pages/"

	addr := flag.String("addr", DefaultAddr, "Listen host/port address of web server.")
	name := flag.String("name", DefaultName, "Name of this wiki.")
	pass := flag.String("pass", DefaultPass, "Password for editing pages. If no password is given, editing is disabled.")
	dir := flag.String("dir", DefaultDir, "Directory of pages markdown files.")
	flag.Parse()

	passHash, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
	if err != nil {
		panic("Could not hash given password.")
	}

	wiki := Wiki{
		Name:     *name,
		PassHash: string(passHash),
		Editable: *pass != DefaultPass,
		Dir:      *dir,
	}

	http.HandleFunc("/e/", wiki.editPage)
	http.HandleFunc("/", wiki.showPage)

	if *pass == "" {
		log.Println("No password was given (with -pass argument), therefore page editing is disabled.")
	}

	log.Printf("Listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
