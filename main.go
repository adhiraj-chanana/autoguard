package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strings"
)

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
}

type CommitResponse struct {
	CommitID string  `json:"commit_id"`
	Status   string  `json:"status"`
	Issues   []Issue `json:"issues"`
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
				})
			}
		}
	}

	// Lint check (flag "print(" usage)
	for _, file := range req.Files {
		lines := strings.Split(file.Content, "\n")
		for i, line := range lines {
			if strings.Contains(line, "print(") {
				issues = append(issues, Issue{
					Type:     "lint",
					Filename: file.Filename,
					Line:     i + 1,
					Message:  "Avoid print statements in production code",
				})
			}
		}
	}

	status := "pass"
	if len(issues) > 0 {
		status = "fail"
	}

	response := CommitResponse{
		CommitID: req.CommitID,
		Status:   status,
		Issues:   issues,
	}

	c.JSON(http.StatusOK, response)
}

func main() {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// New analyze-commit endpoint
	r.POST("/analyze-commit", analyzeCommit)

	r.Run(":8080")
}
