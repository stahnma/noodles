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
    minStars := flag.Int("stars", 1000, "Minimum number of stars a repository must have")
    flag.Parse()

    if flag.NArg() < 1 {
        fmt.Fprintf(os.Stderr, "Usage: %s -stars=<minimum_stars> \"<search_string>\"\n", os.Args[0])
        os.Exit(1)
    }
    searchString := flag.Arg(0)

    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        fmt.Println("GITHUB_TOKEN is not set")
        os.Exit(1)
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

    query := fmt.Sprintf("%s in:readme stars:>%d", searchString, *minStars)

    // Iterate through all available pages
    for {
        result, resp, err := client.Search.Repositories(ctx, query, opts)
        if err != nil {
            fmt.Printf("Error during GitHub search: %s\n", err)
            return
        }

        for _, repo := range result.Repositories {
            if *repo.StargazersCount < *minStars {
                continue
            }
            fmt.Println(*repo.FullName)
        }

        if resp.NextPage == 0 {
            break
        }
        opts.Page = resp.NextPage
    }
}

