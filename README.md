# kagura

A Discord App that provides Arcaea-related functions.

## Development requirements

1. [Go](https://go.dev/) (>= 1.23.4)
2. [Python 3](https://www.python.org/) (Optionally used in development, not required when running)

## Running the app

Kagura can be started by simply running `go run main.go` from the repository's root folder. This requires Go to be installed. More convenient ways to run the app may emerge in the future.

However, before the app can be successfully started, some environment variables need to be configured. See [Configuration](#configuration) for more details.

## Configuration

Kagura is mainly configured via environment variables.

| Environment Variable | Required? |                                                             |
| -------------------- | --------- | ----------------------------------------------------------- |
| KAGURA_TOKEN         | Yes       | Sets the authentication token for the app.                  |
| KAGURA_PREFIX        | No        | Sets the command prefix for text commands. Defaults to `~`. |
| KAGURA_DBPATH        | No        | Sets the SQLite database path. Defaults to `kagura.db`.     |

## Credits

Kagura uses song and chart data obtained from the [Arcaea Fandom wiki](https://arcaea.fandom.com/). Several formulas (e.g. rating and step calculation) are also implemented referring to the information available on the wiki.

This project depends on several libraries, such as [diamondburned/arikawa](https://github.com/diamondburned/arikawa) and [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3). See the [dependency graph](https://github.com/lilacse/kagura/network/dependencies) for more!
