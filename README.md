# Noodles 
---

Why is this named noodles?  Because, I said I would noodle on it.

# GitHub Repository Search Tool

This tool allows you to search GitHub repositories to find those that contain a specific string in their README files and meet a minimum star requirement.

## Requirements

- Go (1.16 or later)
- A GitHub Personal Access Token
- Make

If you want this to run in a lambda, then you you'll want:
- AWSCLI2
- jq (perhaps for debugging)


## Installation

First, clone the repository or download the source code. Then navigate to the directory containing the source code.

## Configuration

Before running the script, you must set your GitHub Personal Access Token as an environment variable:

    export GITHUB_TOKEN="your_github_token_here"

Ensure this token is set in your shell environment where you plan to run the script.

## Building

     flox activate
     go mod tidy
     go build .

## Running as a  lambda

The `Makefile` contains targets to set up or update the code for the lambda.

:warning: You will still need something to trigger the lambda, such as an API
Gateway. Setting that up not covered in this README.

To update the lambda: `make aws`

To create the initial lambda `make aws-init`


## Usage

The tool is executed from the command line, where you can specify the search string and the minimum number of stars:

    ./noodles -stars=<minimum_stars> "<search_string>"

### Parameters

- `<minimum_stars>`: The minimum number of stars a repository must have to be included in the search results. This is optional and defaults to 1000 if not specified.
- `<search_string>`: The string to search for in the README files of repositories.

### Examples

To search for repositories that include "nvm use" in their README and have at least 1500 stars:

    ./noodles -stars=1500 "nvm use"

To search for repositories with the default star setting (1000 stars) that include "pip install":

    ./noodles "pip install"

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues to suggest improvements or add new features.

## License

WTFPL
