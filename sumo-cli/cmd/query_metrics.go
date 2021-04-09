package cmd

import (
	"context"
	"github.com/relvacode/iso8601"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/sumologic/go-sumologic/sumologic/api"
	"github.com/sumologic/go-sumologic/sumologic/util"
	"io/ioutil"
	"os"
	"time"
)

type queryMetricsCmd struct {
	cmd               *cobra.Command
	metricsapi        bool
	csv               bool
	query             string
	queryPath         string
	from              string
	to                string
	timezone          string
	stepSeconds       int
	desiredDataPoints int32
	pollInterval      int64
}

func newQueryMetricsCmd() *queryMetricsCmd {
	result := &queryMetricsCmd{}

	cmd := &cobra.Command{
		Use:   "metrics [...]",
		Short: "Query metrics",
		Long:  `Run a metrics query and output the results`,
		Run:   result.run,
	}
	cmd.Flags().BoolVarP(&result.metricsapi, "metricsapi", "", false, "Use unpublished Metrics API")
	cmd.Flags().BoolVarP(&result.csv, "csv", "", false, "Output as CSV")
	cmd.Flags().StringVarP(&result.query, "query", "q", "", "The metrics query")
	cmd.Flags().StringVarP(&result.queryPath, "file", "", "", "A file containing a metrics query")
	cmd.Flags().StringVarP(&result.from, "from", "f", "", "The ISO 8601 date and time of the time range to start the query. Can also be milliseconds since epoch.")
	cmd.Flags().StringVarP(&result.to, "to", "t", "", "The ISO 8601 date and time of the time range to end the query. Can also be milliseconds since epoch.")
	cmd.Flags().StringVarP(&result.timezone, "timezone", "z", "Etc/UTC", "The time zone if from/to is not in milliseconds. See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for a list of time zone codes.")
	cmd.Flags().IntVarP(&result.stepSeconds, "step", "s", 0, "Desired time quantization step in seconds (when using 'metricsapi'.")
	//cmd.Flags().Int32VarP(&result.desiredDataPoints, "desired-data-points", "p", 0, "The desired number of data points per time series.")
	cmd.Flags().Int64VarP(&result.pollInterval, "polling-interval", "i", 100, "The status polling interval in milliseconds.")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")
	result.cmd = cmd

	return result
}

//TODO error handling

func (c *queryMetricsCmd) run(cmd *cobra.Command, args []string) {
	log.Debug().Str("query", c.query).Msg("queryMetricsCmd run")

	//TODO Validate timezone

	startTime, err := iso8601.ParseString(c.from)
	if err != nil {
		log.Fatal().Str("from", c.from).Err(err)
	}
	startTimeMillis := startTime.UnixNano() / int64(time.Millisecond)
	endTime, err := iso8601.ParseString(c.to)
	if err != nil {
		log.Fatal().Str("from", c.from).Err(err)
	}
	endTimeMillis := endTime.UnixNano() / int64(time.Millisecond)
	log.Debug().Str("from", c.from).Str("to", c.to).
		Str("startTime", startTime.String()).Str("endTime", endTime.String()).
		Int64("startTimeMillis", startTimeMillis).Int64("endTimeMills", endTimeMillis).
		Msg("Parsed timestamps")

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
	transport := api.BasicAuthTransport{Username: AccessId, Password: AccessKey}
	client := api.NewClient(Endpoint, transport.Client())
	ctx := context.Background()

	// Use one of the two possible APIs
	from := util.FromISO8601String(c.from)
	to := util.FromISO8601String(c.to)
	if c.metricsapi {
		//TODO turn into time
		_, t, _ := util.MetricsAPIQuery(client, ctx, c.query, from, to, c.stepSeconds)
		util.OutputTable(os.Stdout, t, c.csv)
	} else {
		_, _, t, _, errors, err := util.DashboardSearch(
			client, ctx, "metrics", c.query, from, to, c.desiredDataPoints, c.pollInterval)
		if err != nil {
			log.Fatal().Err(err).Msg("Error running dashboard search")
		}
		if errors != nil {
			log.Warn().Msgf("Errors in dashboard search result: %v", errors)
		}
		util.OutputTable(os.Stdout, t, c.csv)
	}
}
