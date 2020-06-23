package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/mingcheng/pidfile"
	"github.com/BrandYourself/galera-healthcheck/healthcheck"
	. "github.com/BrandYourself/galera-healthcheck/logger"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var (
		healthchecker         *healthcheck.Healthchecker
		serverPort             int
		mysqlUser              string
		mysqlPassword          string
		availableWhenDonor     bool
		availableWhenReadOnly  bool
		pidfilePath            string
	)

	app := &cli.App{
		Name: "galera-healthcheck",
		Usage: "A lightweight web server to report the health of a node in a Galera cluster",
		Authors: []*cli.Author{
			&cli.Author{ Name: "Stefan Schimanski" },
			&cli.Author{ Name: "Brendan Clougherty", Email: "bclougherty@brand-yourself.com" },
		},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name: "port, P",
				Usage: "Specifies the port of the healthcheck server",
				Value: 8080,
				Destination: &serverPort,
			},
			&cli.StringFlag{
				Name: "user, u",
				Usage: "Specifies the MySQL user to connect as",
				EnvVars: []string{"MYSQL_USER"},
				Destination: &mysqlUser,
				Required: true,
			},
			&cli.StringFlag{
				Name: "password, p",
				Usage: "Specifies the MySQL password to connect with",
				EnvVars: []string{"MYSQL_PASSWORD"},
				Destination: &mysqlPassword,
				Required: true,
			},
			&cli.BoolFlag{
				Name: "availWhenDonor, d",
				Usage: "Specifies if the healthcheck allows availability when in donor state",
				Value: true,
				Destination: &availableWhenDonor,
			},
			&cli.BoolFlag{
				Name: "availWhenReadOnly, r",
				Usage: "Specifies if the healthcheck allows availability when in read only mode",
				Value: false,
				Destination: &availableWhenReadOnly,
			},
			&cli.StringFlag{
				Name: "pidfile, i",
				Usage: "Location for the pidfile",
				Value: "/var/run/galera-healthcheck.pid",
				Destination: &pidfilePath,
			},
		},
		Action: func(c *cli.Context) error {
			if pid, err := pidfile.New(pidfilePath); err != nil {
				return cli.Exit(fmt.Sprintf("Error creating pid file: %+v", err), 1)
			} else {
				defer pid.Remove()
			}

			db, _ := sql.Open("mysql", fmt.Sprintf("%s:%s@/", mysqlUser, mysqlPassword))
			config := healthcheck.HealthcheckerConfig{
				availableWhenDonor,
				availableWhenReadOnly,
			}

			healthchecker = healthcheck.New(db, config)

			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				result, msg := healthchecker.Check()
				if result != nil && result.Healthy {
					w.WriteHeader(http.StatusOK)
				} else if result != nil && !result.Healthy {
					w.WriteHeader(http.StatusServiceUnavailable)
				} else {
					w.WriteHeader(http.StatusContinue)
				}

				fmt.Fprintf(w, "Galera Cluster Node status: %s", msg)
				LogWithTimestamp(msg)
			})

			serverPort := c.Int("port")
			http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
