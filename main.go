package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

// GitHub search function
func searchGitHub(searchString string, minStars int) ([]string, error) {
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

	var results []string
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
			results = append(results, *repo.FullName)
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

func main() {
	// Check if running in Lambda environment
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(handler)
		return
	}

	// CLI logic
	minStars := flag.Int("stars", 1000, "Minimum number of stars a repository must have")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s -stars=<minimum_stars> \"<search_string>\"\n", os.Args[0])
		os.Exit(1)
	}
	searchString := flag.Arg(0)

	results, err := searchGitHub(searchString, *minStars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	for _, repo := range results {
		fmt.Println(repo)
	}
}
