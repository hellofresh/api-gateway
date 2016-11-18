package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/hellofresh/ginger-middleware/mongodb"
	"github.com/hellofresh/ginger-middleware/nice"
	"github.com/hellofresh/janus/config"
	"gopkg.in/alexcesaro/statsd.v2"
	"gopkg.in/redis.v3"
)

var APILoader = APIDefinitionLoader{}

// initializeDatabase initializes a DB connection
func initializeDatabase(dsn string) *mongodb.DatabaseAccessor {
	accessor, err := mongodb.InitDB(dsn)
	if err != nil {
		log.Fatalf("Couldn't connect to the mongodb database: %s", err.Error())
	}

	return accessor
}

// initializeRedis initializes a Redis connection
func initializeRedis(dsn string) *redis.Client {
	log.Debugf("Trying to connect to redis instance: %s", dsn)
	return redis.NewClient(&redis.Options{
		Addr: dsn,
	})
}

// Initializes new StatsD client if it enabled
func initializeStatsd(dsn, prefix string) *statsd.Client {
	var options []statsd.Option

	if len(dsn) == 0 {
		options = append(options, statsd.Mute(true))
	} else {
		options = append(options, statsd.Address(dsn))
	}

	if len(prefix) > 0 {
		options = append(options, statsd.Prefix(prefix))
	}

	client, err := statsd.New(options...)

	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"dsn":    dsn,
				"prefix": prefix,
			}).
			Warning("An error occurred while connecting to StatsD. Client will be muted.")
	}

	return client
}

//loadAPIEndpoints register api endpoints
func loadAPIEndpoints(router *gin.Engine, apiManager *APIManager) {
	log.Debug("Loading API Endpoints")

	handler := AppsAPI{apiManager}
	group := router.Group("/apis")
	{
		group.GET("/", handler.Get())
		group.POST("/", handler.Post())
		group.GET("/:id", handler.GetBy())
		group.PUT("/:id", handler.PutBy())
		group.DELETE("/:id", handler.DeleteBy())
	}
}

func main() {
	log.SetOutput(os.Stderr)

	config, err := config.LoadEnv()
	if nil != err {
		log.Panic(err.Error())
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(nice.Recovery(recoveryHandler))

	if config.Debug {
		log.SetLevel(log.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		log.SetLevel(log.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	}

	accessor := initializeDatabase(config.DatabaseDSN)
	router.Use(mongodb.Middleware(accessor))

	redisStorage := initializeRedis(config.StorageDSN)
	defer redisStorage.Close()

	statsdClient := initializeStatsd(config.StatsdDSN, config.StatsdPrefix)
	defer statsdClient.Close()

	apiManager := NewAPIManager(router, redisStorage, accessor, statsdClient)
	apiManager.Load()
	loadAPIEndpoints(router, apiManager)

	router.Run(fmt.Sprintf(":%v", config.Port))
}
