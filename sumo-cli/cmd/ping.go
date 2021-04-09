package cmd

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/sumologic/go-sumologic/sumologic/api"
)

type pingCmd struct {
	cmd *cobra.Command
}

func newPingCmd() *pingCmd {
	result := &pingCmd{}

	cmd := &cobra.Command{
		Use:   "ping [...]",
		Short: "Validates endpoint and credentials",
		Long:  "Validates endpoint and credentials",
		Run:   result.run,
	}
	result.cmd = cmd

	return result
}

func (c *pingCmd) run(cmd *cobra.Command, args []string) {

	client := api.NewClient(Endpoint, Transport.Client())
	ctx := context.Background()

	_, _, err := client.Users.Users(ctx, 1)
	if err != nil {
		log.Warn().Err(err).Msg("Error getting users")
		fmt.Printf("No success: %s\n", err)
	} else {
		fmt.Printf("Pong! (%s %s)\n", Profile, Endpoint)
	}
}
