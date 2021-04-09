package cmd

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/sumologic/go-sumologic/sumologic/api"
	"github.com/sumologic/go-sumologic/sumologic/util"
	"io/ioutil"
	"os"
)

type queryLogsCmd struct {
	cmd               *cobra.Command
	searchjob         bool
	csv               bool
	query             string
	queryPath         string
	from              string
	to                string
	timezone          string
	desiredDataPoints int32
	pollInterval      int64
}

//TODO error handling

func newQueryLogsCmd() *queryLogsCmd {
	result := &queryLogsCmd{}

	cmd := &cobra.Command{
		Use:   "logs [...]",
		Short: "Query logs",
		Long:  `Run a logs query and output the results`,
		Run:   result.run,
	}
	cmd.Flags().BoolVarP(&result.searchjob, "searchjob", "", false, "Use public SearchJob API")
	cmd.Flags().BoolVarP(&result.csv, "csv", "", false, "Output as CSV")
	cmd.Flags().StringVarP(&result.query, "query", "q", "", "The logs query")
	cmd.Flags().StringVarP(&result.queryPath, "file", "", "", "A file containing a logs query")
	cmd.Flags().StringVarP(&result.from, "from", "f", "", "The ISO 8601 date and time of the time range to start the search. Can also be milliseconds since epoch.")
	cmd.Flags().StringVarP(&result.to, "to", "t", "", "The ISO 8601 date and time of the time range to end the search. Can also be milliseconds since epoch.")
	cmd.Flags().StringVarP(&result.timezone, "timezone", "z", "Etc/UTC", "The time zone if from/to is not in milliseconds. See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for a list of time zone codes.")
	cmd.Flags().Int32VarP(&result.desiredDataPoints, "desired-data-points", "", 0, "The desired number of data points per time series.")
	cmd.Flags().Int64VarP(&result.pollInterval, "polling-interval", "i", 100, "The status polling interval in milliseconds.")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")
	result.cmd = cmd

	return result
}

func (c *queryLogsCmd) run(cmd *cobra.Command, args []string) {
	log.Debug().Str("query", c.query).Msg("queryLogsCmd run")

	//TODO Validate timezone
	//TODO Validate from and to

	// Make sure we have exactly one of query or file
	if c.query == "" && c.queryPath == "" {
		log.Fatal().Str("query", c.query).Str("path", c.queryPath).
			Msg("Need either a file containing a query, or a query")
	}
	if c.query != "" && c.queryPath != "" {
		log.Fatal().Str("query", c.query).Str("path", c.queryPath).
			Msg("Specify only one of file containing a query, or a query")
	}

	// Load query from file if specified
	if c.queryPath != "" {
		content, err := ioutil.ReadFile(c.queryPath)
		if err != nil {
			log.Fatal().Str("path", c.queryPath).Err(err).
				Msg("Error reading file containing query")
		}
		c.query = string(content)
	}

	// Create the API client
	client := api.NewClient(Endpoint, Transport.Client())
	ctx := context.Background()

	// Use one of the two possible APIs
	from := util.FromISO8601String(c.from)
	to := util.FromISO8601String(c.to)
	if c.searchjob {
		//TODO turn into time
		t := util.SearchJob(client, ctx, c.query, c.from, c.to, c.timezone, c.pollInterval)
		util.OutputTable(os.Stdout, t, c.csv)
	} else {
		_, _, t, _, errors, err := util.DashboardSearch(
			client, ctx, "logs", c.query, from, to, c.desiredDataPoints, c.pollInterval)
		if err != nil {
			log.Fatal().Err(err).Msg("Error running dashboard search")
		}
		if errors != nil {
			log.Fatal().Msgf("Errors in dashboard search result: %v", errors)
		}
		util.OutputTable(os.Stdout, t, c.csv)
	}
}
