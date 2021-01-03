package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	_headerContentType  = "Content-Type"
	_headerCacheControl = "Cache-Control"
	_mimeTextHTML       = "text/html"
	_queryParamGoGet    = "go-get"
	_templateSource     = `<html>
    <head>
        <meta name="go-import" content="{{ .VanityDomain }}{{ .VanityPath }} git {{ .GitHubOrgURL }}/{{ .GitRepository }}">
        <meta name="go-source" content="{{ .VanityDomain }}{{ .VanityPath }}     {{ .GitHubOrgURL }}/{{ .GitRepository }} {{ .GitHubOrgURL }}/{{ .GitRepository }}/tree/{{ .GitHubBranch }}{/dir} {{ .GitHubOrgURL }}/{{ .GitRepository }}/blob/{{ .GitHubBranch }}{/dir}/{file}#L{line}">
    </head>
</html>
`
	_valueCacheControl = "public, max-age=86400"
)

var (
	tmpl = template.Must(template.New("vanity-url").Parse(_templateSource))
)

type templateContext struct {
	GitHubBranch  string
	GitHubOrgURL  string
	GitRepository string
	VanityDomain  string
	VanityPath    string
}

func main() {
	log.Print("starting server...")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	domain := os.Getenv("VANITY_DOMAIN")
	if domain == "" {
		log.Fatal(fmt.Errorf("required env variable: VANITY_DOMAIN"))
	}

	githubOrg := os.Getenv("GITHUB_ORG")
	if githubOrg == "" {
		log.Fatal(fmt.Errorf("required env variable: GITHUB_ORG"))
	}

	githubOrgURL := fmt.Sprintf("https://github.com/%s", githubOrg)

	githubBranch := os.Getenv("GITHUB_BRANCH")
	if githubBranch == "" {
		log.Fatal(fmt.Errorf("required env variable: GITHUB_BRANCH"))
	}

	http.HandleFunc("/", vanityHandler(domain, githubOrgURL, githubBranch))

	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}

func vanityHandler(domain, githubOrgURL, githubBranch string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Query().Get(_queryParamGoGet) != "1" || r.URL.Path == "/" {
			http.Redirect(w, r, githubOrgURL, http.StatusMovedPermanently)
			return
		}

		// fixme: /service/foundation => service-foundation
		// fixme: /service/api-authorizer => service-api-authorizer
		pathParts := strings.Split(r.URL.Path, "/")[1:]
		tmplCtx := templateContext{
			GitHubBranch:  githubBranch,
			GitHubOrgURL:  githubOrgURL,
			GitRepository: strings.Join(pathParts, "-"),
			VanityDomain:  domain,
			VanityPath:    r.URL.Path,
		}

		buf := bytes.NewBuffer([]byte{})
		if err := tmpl.Execute(buf, tmplCtx); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		w.Header().Set(_headerContentType, _mimeTextHTML)
		w.Header().Set(_headerCacheControl, _valueCacheControl)
		w.WriteHeader(http.StatusOK)

		if _, err := buf.WriteTo(w); err != nil {
			log.Fatal(err)
		}
	}
}
