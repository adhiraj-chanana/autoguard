package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strings"
	"github.com/gin-contrib/cors"
	"log" 
	"time"
)
type CommitWithIssues struct {
	CommitID  string  `json:"commit_id"`
	RepoURL   string  `json:"repo_url"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
	Issues    []Issue `json:"issues"`
}

type HistoryResponse struct {
	Commits []CommitWithIssues `json:"commits"`
}
type File struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type CommitRequest struct {
	CommitID string `json:"commit_id"`
	RepoURL  string `json:"repo_url"`
	Files    []File `json:"files"`
}

type Issue struct {
	Type     string `json:"type"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	Retries  int    `json:"retries"`
}

type CommitResponse struct {
	CommitID string  `json:"commit_id"`
	Status   string  `json:"status"`
	Issues   []Issue `json:"issues"`
}

// retryWithBackoff retries a function with exponential backoff delays.
func retryWithBackoff(attempts int, fn func() error) error {
	var err error
	delay := time.Second // start at 1 second

	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil // success
		}

		if i < attempts-1 {
			time.Sleep(delay)
			delay *= 2 // exponential backoff
		}
	}

	return errors.New("all retries failed: " + err.Error())
}

func analyzeCommit(c *gin.Context) {
	var req CommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var issues []Issue

	// Secret scanner (regex for "password" or "API_KEY")
	secretRegex := regexp.MustCompile(`(?i)(password|api[_-]?key)`)
	for _, file := range req.Files {
		lines := strings.Split(file.Content, "\n")
		for i, line := range lines {
			if secretRegex.MatchString(line) {
				issues = append(issues, Issue{
					Type:     "secret",
					Filename: file.Filename,
					Line:     i + 1,
					Message:  "Possible hardcoded secret detected",
					Retries:  0,
				})
			}
		}
	}

	// Lint check with retry logic
	for _, file := range req.Files {
		lines := strings.Split(file.Content, "\n")
		for i, line := range lines {
			retries := 3
			retryErr := retryWithBackoff(retries, func() error {
				if strings.Contains(line, "print(") {
					return errors.New("lint violation")
				}
				return nil
			})

			if retryErr != nil {
				issues = append(issues, Issue{
					Type:     "lint",
					Filename: file.Filename,
					Line:     i + 1,
					Message:  fmt.Sprintf("Avoid print statements in production code"),
					Retries:  retries,
				})
			}
		}
	}

	// Final status
	status := "pass"
	if len(issues) > 0 {
		status = "fail"
	}
	// Save commit result
	_, err := db.Exec(
		"INSERT INTO commits (commit_id, repo_url, status) VALUES (?, ?, ?)",
		req.CommitID, req.RepoURL, status,
	)
	if err != nil {
		log.Println("Error saving commit:", err)
	}

	// Save issues
	for _, issue := range issues {
		_, err := db.Exec(
			"INSERT INTO issues (commit_id, type, filename, line, message, retries) VALUES (?, ?, ?, ?, ?, ?)",
			req.CommitID, issue.Type, issue.Filename, issue.Line, issue.Message, issue.Retries,
		)
		if err != nil {
			log.Println("Error saving issue:", err)
		}
	}


	response := CommitResponse{
		CommitID: req.CommitID,
		Status:   status,
		Issues:   issues,
	}

	c.JSON(http.StatusOK, response)
}
func getHistory(c *gin.Context) {
	// Optional limit query param (default 5)
	limit := c.DefaultQuery("limit", "5")

	rows, err := db.Query(
		"SELECT commit_id, repo_url, status, timestamp FROM commits ORDER BY id DESC LIMIT ?",
		limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch history"})
		return
	}
	defer rows.Close()

	var commits []CommitWithIssues

	for rows.Next() {
		var commit CommitWithIssues
		if err := rows.Scan(&commit.CommitID, &commit.RepoURL, &commit.Status, &commit.Timestamp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse commits"})
			return
		}

		// Fetch issues for this commit
		issueRows, err := db.Query(
			"SELECT type, filename, line, message, retries FROM issues WHERE commit_id = ?",
			commit.CommitID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch issues"})
			return
		}

		var issues []Issue
		for issueRows.Next() {
			var issue Issue
			if err := issueRows.Scan(&issue.Type, &issue.Filename, &issue.Line, &issue.Message, &issue.Retries); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse issues"})
				return
			}
			issues = append(issues, issue)
		}
		issueRows.Close()

		commit.Issues = issues
		commits = append(commits, commit)
	}

	c.JSON(http.StatusOK, HistoryResponse{Commits: commits})
}


func main() {
	initDB()
	defer db.Close()
	r := gin.Default()
	r.Use(cors.Default())


	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/history", getHistory)
	r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type"},
    AllowCredentials: true,
}))


	// Analyze commit
	r.POST("/analyze-commit", analyzeCommit)

	// Start server
	r.Run(":8080")
}
