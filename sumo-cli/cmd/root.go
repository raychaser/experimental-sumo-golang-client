package cmd

import (
	"fmt"
	"github.com/henvic/httpretty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/sumologic/go-sumologic/sumologic/api"
	"github.com/sumologic/sumo-cli/util"
	"net/http"
	"os"
	"time"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"

	consoleLogging bool
	debugLogging   bool
	httpLogging    bool

	CredentialsPath string
	Profile         string
	AccessId        string
	AccessKey       string
	Endpoint        string

	Transport api.BasicAuthTransport
)

func NewRootCmd(v string, c string, d string, b string) *cobra.Command {
	cobra.OnInitialize(initLogging, initTransport)

	version = v
	commit = c
	date = d
	builtBy = b

	result := &cobra.Command{
		Use:   "sumo",
		Short: "A CLI client for Sumo Logic",
		Long:  `...`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Debug().Msg("result run")
		},
	}

	result.PersistentFlags().StringVarP(&CredentialsPath, "credentials", "c", "", "File with Sumo Logic credentials")
	result.PersistentFlags().StringVarP(&Profile, "Profile", "p", "default", "Name of the Profile to use for credentials")
	result.PersistentFlags().BoolVarP(&consoleLogging, "console", "", false, "Enable logging on the console")
	result.PersistentFlags().BoolVarP(&debugLogging, "debug", "", false, "Enable debug logging if --console is enabled")
	result.PersistentFlags().BoolVarP(&httpLogging, "http", "", false, "Enable HTTP logging if --console is enabled")

	result.AddCommand(newVersionCmd())
	result.AddCommand(newPingCmd().cmd)
	result.AddCommand(newQueryCmd())

	return result
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the Sumo Logic CLI client",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Sumo Logic CLI client %s (commit: %s date: %s, builtBy: %s)\n",
				version, commit, date, builtBy)
		},
	}
}

func init() {
	log.Debug().Str("version", version).Str("commit", commit).Str("date", date).
		Str("builtBy", builtBy).Msg("root.go init()")
}

func initLogging() {
	//fAll, _ := os.OpenFile("./sumo-cli.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	//consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}}
	//multiWriter := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)
	//logger := zerolog.New(multi).With().Timestamp().Logger()

	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	//TODO Make these into proper Cobra errors
	if debugLogging && !consoleLogging {
		fmt.Printf("Debug logging requires console logging enabled (--console)\n")
		os.Exit(-0)
	}
	if httpLogging && !consoleLogging {
		fmt.Printf("HTTP logging requires console logging enabled (--console)\n")
		os.Exit(0)
	}

	// If console logging is requested, set it up
	if consoleLogging {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
		if debugLogging {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Debug().Msg("Log level set to debug")
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
	} else {

		// No console logging is requested, so only log FATAL and above
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	}

	log.Info().Msg("Sumo Logic CLI client starting")
}

func initTransport() {
	id, key, endpoint, err := util.LoadCredentials(CredentialsPath, Profile)
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading credentials")
	}
	AccessId = id
	AccessKey = key
	Endpoint = endpoint

	if httpLogging {
		log.Debug().Msg("HTTP logging enabled")
		logger := &httpretty.Logger{
			Time:            true,
			TLS:             true,
			RequestHeader:   true,
			RequestBody:     true,
			MaxRequestBody:  1024 * 1024,
			ResponseHeader:  true,
			ResponseBody:    true,
			MaxResponseBody: 1024 * 1024,
			Colors:          true,
			Formatters:      []httpretty.Formatter{&httpretty.JSONFormatter{}},
		}
		Transport = api.BasicAuthTransport{Username: AccessId, Password: AccessKey,
			Transport: logger.RoundTripper(http.DefaultTransport)}
	} else {
		Transport = api.BasicAuthTransport{Username: AccessId, Password: AccessKey}
	}
}
