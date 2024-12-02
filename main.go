package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"golang.org/x/term"
)

// RepositoryResult represents a single repository's details
type RepositoryResult struct {
	Slug  string `json:"slug"`
	Stars int    `json:"stars"`
	URL   string `json:"url"`
}

// GitHub search function
func searchGitHub(searchString string, minStars int) ([]RepositoryResult, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opts := &github.SearchOptions{
		Sort:        "stars",
		Order:       "desc",
		ListOptions: github.ListOptions{PerPage: 100}, // Maximize the results per page
	}

	query := fmt.Sprintf("%s in:readme stars:>%d", searchString, minStars)

	var results []RepositoryResult
	// Iterate through all available pages
	for {
		result, resp, err := client.Search.Repositories(ctx, query, opts)
		if err != nil {
			return nil, fmt.Errorf("error during GitHub search: %s", err)
		}

		for _, repo := range result.Repositories {
			if *repo.StargazersCount < minStars {
				continue
			}
			results = append(results, RepositoryResult{
				Slug:  *repo.FullName,
				Stars: *repo.StargazersCount,
				URL:   *repo.HTMLURL,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return results, nil
}

// Lambda handler function
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	searchString := request.QueryStringParameters["search"]
	minStarsStr := request.QueryStringParameters["stars"]

	if searchString == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "search parameter is required"}, nil
	}

	minStars, err := strconv.Atoi(minStarsStr)
	if err != nil || minStars < 0 {
		minStars = 1000 // Default value
	}

	results, err := searchGitHub(searchString, minStars)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Error: %s", err)}, nil
	}

	// Log headers for debugging
	log.Printf("Headers: %+v", request.Headers)

	// Normalize headers to lowercase
	acceptHeader := ""
	for k, v := range request.Headers {
		if strings.ToLower(k) == "accept" {
			acceptHeader = strings.ToLower(v)
			break
		}
	}

	// Handle plain text output
	if strings.Contains(acceptHeader, "text/plain") {
		var plainTextBody strings.Builder
		for _, repo := range results {
			plainTextBody.WriteString(fmt.Sprintf("%s (Stars: %d, URL: %s)\n", repo.Slug, repo.Stars, repo.URL))
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			Body: plainTextBody.String(),
		}, nil
	}

	// Default to JSON output
	jsonResponse, err := json.Marshal(results)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Error marshaling JSON: %s", err)}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonResponse),
	}, nil
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func main() {
	// Check if running in Lambda environment
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(handler)
		return
	}

	// CLI logic
	minStars := flag.Int("stars", 1000, "Minimum number of stars a repository must have")
	outputJSON := flag.Bool("json", false, "Output results in JSON format")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s -stars=<minimum_stars> [-json] \"<search_string>\"\n", os.Args[0])
		os.Exit(1)
	}
	searchString := flag.Arg(0)

	results, err := searchGitHub(searchString, *minStars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if *outputJSON {
		var jsonOutput []byte
		if isTTY() {
			// Pretty print JSON if running in a TTY
			jsonOutput, err = json.MarshalIndent(results, "", "  ")
		} else {
			jsonOutput, err = json.Marshal(results)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonOutput))
	} else {
		for _, repo := range results {
			fmt.Printf("%s (Stars: %d, URL: %s)\n", repo.Slug, repo.Stars, repo.URL)
		}
	}
}
