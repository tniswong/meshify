# Meshify Coding Assessment

> The code exercise:
>    In a language of your choosing, use the Twitter API (https://developer.twitter.com/en/docs.html) to gather
>    2000 unique tweets with the hashtag #IoT and output them to a CSV file.  Please implement this solution with
>    concurrency in some way.
>
>    We'll be evaluating on:
>    * Unit tests
>    * Documentation
>    * Code organization

## Run Tests

    $> make deps   # only necessary once
    $> make ensure # only necessary once
    $> make test

## Build the binary

    $> make ensure # only necessary once
    $> make

## Running the binary

    $ ./meshify -k <yourKey> -s <yourSecret> -o out.csv -n 2000 -t IoT,Help

    $ ./meshify --help
    Use the Twitter API to gather 2000 unique tweets with the hashtag #IoT and output them to a CSV file.

    Usage:
      meshify [flags]

    Flags:
      -k, --api-key string      Required. Twitter API Public Key. If unset uses MESHIFY_API_KEY environment variable.
      -s, --api-secret string   Required. Twitter API Secret Key. If unset uses MESHIFY_API_SECRET environment variable.
      -h, --help                help for meshify
      -n, --number int          Number of tweets per hashtag. (default 2000)
      -o, --out string          Output file path for csv formatted output. (default STDOUT)
      -t, --tags strings        Hashtags to query. Note: '#' is a comment character in bash, so use ONLY the tag name rather than the full hashtag i.e. 'IoT' NOT '#IoT'! (default [IoT])

## Make targets

1. `make clean`

    Deletes leftover `.coverprofile` files.

1. `make doc`

    Starts a `godoc` server for this package.

1. `make deps`

    Install all dependent cli's for these make targets. Run this first, at least once!

1. `make ensure`

    Ensure all runtime dependencies are installed properly.

1. `make fmt` or `make format`

    Automatically format all code in this package.

1. `make vet`

    Run `go vet` on all code in this package, excluding dependencies. Exit 0, if successful. Exit 1, if not.

1. `make lint`

    Run `go lint` on all code in this package, excluding dependencies. Exit 0, if successful. Exit 1, if not.

1. `make complexity`

    Generate a complexity report for all code in this package, excluding dependencies. Exit 0, if reported complexity is
    above maximum threshold. Exit 1, if not.

1. `make coverage`

    Generate a coverage report for all code in this package, excluding dependencies. Exit 0, if reported coverage is
    below minimum threshold. Exit 1, if not.

1. `make test`

    Vet, Lint, Test with Coverage, and complexity. Exit 0, if successful. Exit 1, if there is unformatted code, if there
    are lint failures, if there are test failures, if coverage is below the minimum threshold, or if complexity is above
    the maximum threshold.

1. `make build` or `make`

    Build the binary