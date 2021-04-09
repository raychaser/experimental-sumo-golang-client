package util

import (
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
	"os"
	"strings"
)

const (
	defaultCredentialsPath = "/.sumo/credentials"
)

func LoadCredentials(path string, profile string) (string, string, string, error) {
	creds := os.Getenv("CREDENTIALS")
	if creds != "" {
		return fromEnv(creds)
	} else {
		return fromFile(path, profile)
	}
}

func fromEnv(creds string) (string, string, string, error) {
	log.Info().Msg("Using credentials from environment")
	credentials := strings.Split(creds, ":")
	if len(credentials) != 2 {
		log.Fatal().Msg("Splitting $CREDENTIALS by ':' did not result in two values")
	}
	accessId := credentials[0]
	accessKey := credentials[1]
	endpoint := os.Getenv("URL")
	if endpoint == "" {
		log.Fatal().Msg("No $URL specified")
	}
	return accessId, accessKey, endpoint, nil
}

func fromFile(path string, profile string) (string, string, string, error) {
	p := path
	if path == "" {
		log.Info().Str("path", defaultCredentialsPath).
			Msg("Loading credentials from standard location")
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal().Err(err).Msg("Error getting home directory")
		}
		p = home + "/.sumo/credentials"
	}
	log.Info().Str("path", p).Str("profile", profile).Msg("Loading credentials")
	cfg, err := ini.Load(p)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get credentials")
	}
	section, err := cfg.GetSection(profile)
	if err != nil {
		return "", "", "", err
	}
	accessId := section.Key("accessId").String()
	accessKey := section.Key("accessKey").String()
	endpoint := section.Key("endpoint").String()
	return accessId, accessKey, endpoint, nil
}
