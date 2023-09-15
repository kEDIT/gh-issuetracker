package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/v54/github"
	"golang.org/x/exp/slog"
)

type Server struct {
	State     AppState
	apiClient *github.Client
	logger    *slog.Logger
}

type AppState struct {
	Repo    string
	Owner   string
	Options *github.IssueListByRepoOptions
	Issues  []IssueSummary
}

type IssueSummary struct {
	ID        int64
	Title     string
	Link      string // link to github issue page, not the link to the api endpoint
	State     string
	CreatedAt string
	UpdatedAt string
}

// Returns repository owner, repository name, and parsing error
func extractRepoInfo(repoURL string) (string, string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}

	// 'Owner' and 'Repo' should be the second and third elements in 'u.Path' respectively
	parts := strings.Split(u.Path, "/")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid URL: %s", repoURL)
	}

	return parts[1], parts[2], nil
}

func (s *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "index.html", s.State.Issues)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "unable to execute template")
		http.Error(w, "failed to execute template", http.StatusInternalServerError)
	}
}

func (s *Server) ListIssuesHandler(w http.ResponseWriter, r *http.Request) {

	// Clear the `Issues slice`
	s.State.Issues = s.State.Issues[:0]

	ctx := r.Context()
	tmpl, err := template.ParseFiles("./static/index.html")
	if err != nil {
		s.logger.ErrorContext(ctx, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	// Split repository URL into parts to identify the repository name and owner
	repoURL := r.FormValue("repoURL")
	owner, repo, err := extractRepoInfo(repoURL)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to parse input URL: ", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	// NOTE: AppState doesn't really need to store owner or repo information, it's just there to illustrate
	// how I would hold application state in a larger project
	s.State.Owner = owner
	s.State.Repo = repo

	state := r.FormValue("issueState")
	sortBy := r.FormValue("sortBy")
	direction := r.FormValue("direction")

	opts := &github.IssueListByRepoOptions{
		State:     state,
		Sort:      sortBy,
		Direction: direction,
	}

	opts.ListOptions = github.ListOptions{
		Page: 10,
	}

	issues, _, err := s.apiClient.Issues.ListByRepo(ctx, s.State.Owner, s.State.Repo, opts)

	if err != nil {
		s.logger.ErrorContext(ctx, err.Error())
	} else {
		for _, issue := range issues {

			i := IssueSummary{
				ID:        issue.GetID(),
				Title:     issue.GetTitle(),
				Link:      issue.GetHTMLURL(),
				State:     issue.GetState(),
				CreatedAt: issue.GetCreatedAt().Local().Format("02 Jan 06 15:04 MST"),
				UpdatedAt: issue.GetUpdatedAt().Local().Format("02 Jan 06 15:04 MST"),
			}
			s.State.Issues = append(s.State.Issues, i)
			tmpl.ExecuteTemplate(w, "issue-list-element", i)
		}
	}
}

//go:embed static
var fs embed.FS

// FIXME: should a global template be used here? Seems bad.
var tmpl = template.Must(template.ParseFS(fs, "static/*.html", "static/*.css", "static/*.js"))

func main() {

	logHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logHandler)

	server := Server{
		State:     AppState{Issues: []IssueSummary{}},
		apiClient: github.NewClient(nil),
		logger:    logger,
	}

	http.HandleFunc("/", server.RootHandler)
	http.Handle("/static/", http.FileServer(http.FS(fs)))
	http.HandleFunc("/list", server.ListIssuesHandler)

	log.Fatal(http.ListenAndServe(":8090", nil))
}
