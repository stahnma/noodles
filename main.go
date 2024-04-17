package main

import (
    "context"
    "flag"
    "fmt"
    "os"

    "github.com/google/go-github/v39/github"
    "golang.org/x/oauth2"
)

func main() {
    // Define command line flags
    minStars := flag.Int("stars", 1000, "Minimum number of stars a repository must have")
    flag.Parse() // Parse the flags

    // Check for command line arguments after flags
    if flag.NArg() < 1 {
        fmt.Printf("Usage: %s -stars=1000 <search-string>\n", os.Args[0])
        os.Exit(1)
    }
    searchString := flag.Arg(0) // Get the search string from the first non-flag argument

    // Read the GitHub token from the environment variable
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        fmt.Println("GITHUB_TOKEN is not set")
        os.Exit(1)
    }

    // Set up OAuth2 access with the token
    ctx := context.Background()
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(ctx, ts)

    // Create a GitHub client
    client := github.NewClient(tc)

    // Define search options
    opts := &github.SearchOptions{
        Sort:  "stars",  // Sort by stars
        Order: "desc",   // Order in descending order
    }

    // Format the search query with the custom search string and star filter
    query := fmt.Sprintf("%s in:readme stars:>%d", searchString, *minStars)
    result, _, err := client.Search.Repositories(ctx, query, opts)
    if err != nil {
        fmt.Printf("Error during GitHub search: %s\n", err)
        return
    }

    // Process each repository found
    for _, repo := range result.Repositories {
        // Output the repository's full name (organization/repository)
        fmt.Println(*repo.FullName)
    }
}


