package main

import (
	"os"
	"time"

	"github.com/jirevwe/cascade/internal/pkg/config"
	"github.com/jirevwe/cascade/internal/pkg/datastore"
	"github.com/jirevwe/cascade/internal/pkg/listen"
	"github.com/jirevwe/cascade/internal/pkg/queue"
	"github.com/jirevwe/cascade/internal/pkg/server"
	"github.com/jirevwe/cascade/internal/pkg/tasks"
	"github.com/jirevwe/cascade/internal/pkg/util"
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
			log.Fatal("server: failed to set env - ", err)
		}

		err = config.LoadConfig(configFile)
		if err != nil {
			log.Fatal("server: failed to load config - ", err)
		}

		c, err := config.Get()
		if err != nil {
			log.Fatal("server: failed to set env - ", err)
		}

		rdb, err := queue.NewRedis(c.RedisDsn)
		if err != nil {
			log.Fatal("server: failed to connect to redis - ", err)
		}

		queueNames := map[string]int{string(util.DeleteEntityQueue): 10}
		opts := queue.QueueOptions{
			Names:        queueNames,
			RedisClient:  rdb,
			RedisAddress: c.RedisDsn,
		}
		q := queue.NewQueue(opts)

		// register worker.
		consumer, err := tasks.NewConsumer(q)
		if err != nil {
			log.Fatal("server: failed to create consumer - ", err)
		}

		db, err := datastore.New(c)
		if err != nil {
			log.Fatal("server: failed to connect to db - ", err)
		}

		consumer.RegisterHandlers(util.DeleteEntityTask, tasks.DeleteEntity(db, rdb))

		//start worker
		log.Info("server: starting workers...")
		consumer.Start()

		// starts mongodb change stream and goroutines
		listen.New(c, db, rdb, q)

		srv := server.NewServer(c.Port)
		srv.SetHandler(q.(*queue.RedisQueue).Monitor())

		log.Infof("server: running on port %v", c.Port)
		srv.Listen()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&configFile, "config", "./config.json", "Configuration file")
}
