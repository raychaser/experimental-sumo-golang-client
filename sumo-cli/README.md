# WARNING

This is an initial **EXPERIMENTAL** attempt at creating both
Sumo Logic API bindings for Golang as well as an actual CLI
to interact with the Sumo Logic service.

Please consider this "my first Go project"––the author is NOT
a professional Go programmer. However, an attempt was made to
follow best practices for building API and CLI clients.

At this point this code is not meant for production use. This
code is **NOT** an official product of Sumo Logic.


# Sumo CLI

This is a CLI to interact with the Sumo Logic API.

## Table Of Contents

- [Sumo CLI](#sumo-cli)
  - [Table Of Contents](#table-of-contents)
  - [Setup](#setup)
  - [General Structure](#general-structure)
  - [Commands](#commands)
    - [Query](#query)
      - [Logs](#logs)


## Setup

`sumo-cli` needs to know how to connect to Sumo Logic. For this, a 
simple "INI"-style configuration file is used, inspired by how this 
works for AWS CLI tools. By default `sumo-cli` will look for the 
configuration in `~/.sumo/credentials` but it is possible to 
specify another location using `-c`/`--credentials`. Each configuration
file can contain one or multiple profiles. By default, `sumo-cli` 
will look for profile `default`. A different profile can be specified
using `-p`/`--profile`.

Example configuration file, assume location is `~/.sumo/credentials`:

```ini
[default]
accessId=sXXXXXXXXXXXXX
accessKey=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
endpoint=https://api.sumologic.com/api/
[australia]
accessId=ssXXXXXXXXXXXXX
accessKey=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
endpoint=https://api.au.sumologic.com/api/
[deutschland]
accessId=asXXXXXXXXXXXXX
accessKey=SXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
endpoint=https://api.de.sumologic.com/api/
```

Here, we have defined three profiles: `default` (with endpoint US1 API), 
`australia` (endpoint AU API) and `deutschland` (endpoint DE API).

For information on endpoints please see [Sumo Logic Endpoints and Firewall Security](https://help.sumologic.com/APIs/General-API-Information/Sumo-Logic-Endpoints-and-Firewall-Security). For information on creating API access keys please see [Access Keys](https://help.sumologic.com/Manage/Security/Access-Keys).

To validate configuration is correct for a profile, we can use the `ping` 
command:

```sh
$ sumo-cli ping
Pong! (default https://api.sumologic.com/api/)
```

You can tell `sumo-cli` which profile to use with the `-p/--profile` switch:

```sh
$ sumo-cli -p deutschland ping
Pong! (deutschland https://api.de.sumologic.com/api/)
```

## General Structure

This CLI follows a command/subcommand structure. We can see all the 
supported commands:

```sh
$ sumo-cli --help
Usage:
  sumo [flags]
  sumo [command]

Available Commands:
  help        Help about any command
  ping        Validates endpoint and credentials
  query
  version     Print the version number of the Sumo Logic CLI client

Flags:
  -p, --Profile string       Name of the Profile to use for credentials (default "default")
      --console              Enable logging on the concole
  -c, --credentials string   File with Sumo Logic credentials
  -d, --debug                Enable debug logging
  -h, --help                 help for sumo

Use "sumo [command] --help" for more information about a command.
```

We can then see the subcommands available, for example, `query`:

```sh
$ sumo-cli logs --help
Usage:
  sumo query [command]

Available Commands:
  logs        Query logs
  metrics     Query metrics

Flags:
  -h, --help   help for query

Global Flags:
  -p, --Profile string       Name of the Profile to use for credentials (default "default")
      --console              Enable logging on the concole
  -c, --credentials string   File with Sumo Logic credentials
  -d, --debug                Enable debug logging

Use "sumo query [command] --help" for more information about a command.
```
## Commands

### Query

#### Logs

Query logs in Sumo Logic with this command.

Currently supports only aggregate queries.

```
Run a logs query and output the results

Usage:
  sumo query logs [...] [flags]

Flags:
      --csv                         Output as CSV
      --desired-data-points int32   The desired number of data points per time series.
      --file string                 A file containing a logs query
  -f, --from string                 The ISO 8601 date and time of the time range to start the search. Can also be milliseconds since epoch.
  -h, --help                        help for logs
  -i, --polling-interval int        The status polling interval in milliseconds. (default 100)
  -q, --query string                The logs query
      --searchjob                   Use public SearchJob API
  -z, --timezone string             The time zone if from/to is not in milliseconds. See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for a list of time zone codes. (default "Etc/UTC")
  -t, --to string                   The ISO 8601 date and time of the time range to end the search. Can also be milliseconds since epoch.

Global Flags:
  -p, --Profile string       Name of the Profile to use for credentials (default "default")
      --console              Enable logging on the concole
  -c, --credentials string   File with Sumo Logic credentials
  -d, --debug                Enable debug logging
```

Count the number of logs per minute for the first 15 minutes of today:

```sh
./sumo-cli query logs \
-f "`date +"%Y-%m-%d"`T00:00:00" \
-t "`date +"%Y-%m-%d"`T00:01:00" \
-q 'ERROR | timeslice 1s | count _timeslice | sort + _timeslice'
```

Result:

```
┌────┬───────────────────────────────┬─────────┐
│    │ _TIMESLICE                    │  _COUNT │
├────┼───────────────────────────────┼─────────┤
│  1 │ 2020-11-07 16:00:00 -0800 PST │ 1367509 │
│  2 │ 2020-11-07 16:00:01 -0800 PST │  175280 │
│  3 │ 2020-11-07 16:00:02 -0800 PST │   60503 │
...
│ 58 │ 2020-11-07 16:00:57 -0800 PST │   20123 │
│ 59 │ 2020-11-07 16:00:58 -0800 PST │   15503 │
│ 60 │ 2020-11-07 16:00:59 -0800 PST │   16858 │
└────┴───────────────────────────────┴─────────┘
```

Use `--csv` to get the output formatted as CSV.

`sumo-cli` uses the Dashboard Search API to query log aggregates.
Use `--searchjob` to force the use of the [Search Job API](https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API).
