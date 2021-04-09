module github.com/sumologic/sumo-cli

go 1.14

require (
	github.com/go-openapi/strfmt v0.19.5 // indirect
	github.com/henvic/httpretty v0.0.6
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/relvacode/iso8601 v1.1.0
	github.com/rs/zerolog v1.19.0
	github.com/spf13/cobra v1.0.0
	github.com/sumologic/go-sumologic v0.0.0
	gopkg.in/ini.v1 v1.57.0
)

replace github.com/sumologic/go-sumologic v0.0.0 => ./../go-sumologic
