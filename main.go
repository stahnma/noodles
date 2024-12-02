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
		log.Printf("Missing 'search' parameter")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "search parameter is required",
		}, nil
	}

	minStars, err := strconv.Atoi(minStarsStr)
	if err != nil || minStars < 0 {
		log.Printf("Invalid 'stars' parameter: %v", minStarsStr)
		minStars = 1000 // Default value
	}

	results, err := searchGitHub(searchString, minStars)
	if err != nil {
		log.Printf("Error during GitHub search: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Internal server error",
		}, nil
	}

	// If no results, return 200 with an empty array
	if len(results) == 0 {
		log.Printf("No results found for search: '%s', minStars: %d", searchString, minStars)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: "[]", // Empty JSON array
		}, nil
	}

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
		log.Printf("Error marshaling JSON: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Internal server error",
		}, nil
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

	// Check if stdout is a TTY
	if isTTY() && !*outputJSON {
		// Output only slugs, one per line
		for _, repo := range results {
			fmt.Println(repo.Slug)
		}
		return
	}

	// Handle JSON or full details
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
