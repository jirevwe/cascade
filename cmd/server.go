package main

import (
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jirevwe/cascade/internal/pkg/config"
	"github.com/jirevwe/cascade/internal/pkg/server"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var configFile string

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"s"},
	Short:   "Starts the http server",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&prefixed.TextFormatter{
			TimestampFormat: time.RFC3339,
			DisableColors:   false,
			FullTimestamp:   true,
			ForceFormatting: true,
		})
		log.SetReportCaller(true)

		err := os.Setenv("TZ", "") // Use UTC by default :)
		if err != nil {
			log.Fatal("failed to set env - ", err)
		}

		err = config.LoadConfig(configFile)
		if err != nil {
			log.Fatal("failed to load config - ", err)
		}

		c, err := config.Get()
		if err != nil {
			log.Fatal("failed to set env - ", err)
		}

		srv := server.NewServer(c.Port)
		srv.SetHandler(chi.NewRouter())

		log.Infof("server running on port %v", c.Port)
		srv.Listen()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&configFile, "config", "./config.json", "Configuration file")
}
