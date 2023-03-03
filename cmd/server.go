// Copyright 2023 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/uptimedog/badger/core/controller"
	mid "github.com/uptimedog/badger/core/middleware"
	"github.com/uptimedog/badger/core/service"

	"github.com/drone/envsubst"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/trace"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the badger backend server",
	Run: func(cmd *cobra.Command, args []string) {
		configUnparsed, err := ioutil.ReadFile(config)

		if err != nil {
			panic(fmt.Sprintf(
				"Error while reading config file [%s]: %s",
				config,
				err.Error(),
			))
		}

		configParsed, err := envsubst.EvalEnv(string(configUnparsed))

		if err != nil {
			panic(fmt.Sprintf(
				"Error while parsing config file [%s]: %s",
				config,
				err.Error(),
			))
		}

		viper.SetConfigType("yaml")
		err = viper.ReadConfig(bytes.NewBuffer([]byte(configParsed)))

		if err != nil {
			panic(fmt.Sprintf(
				"Error while loading configs [%s]: %s",
				config,
				err.Error(),
			))
		}

		fys := service.NewFileSystem()

		if viper.GetString("server.log.output") != "stdout" {
			dir, _ := filepath.Split(viper.GetString("server.log.output"))

			// Create dir
			if !fys.DirExists(dir) {
				if err := fys.EnsureDir(dir, 0775); err != nil {
					panic(fmt.Sprintf(
						"Directory [%s] creation failed with error: %s",
						dir,
						err.Error(),
					))
				}
			}

			// Create log file if not exists
			if !fys.FileExists(viper.GetString("server.log.output")) {
				f, err := os.Create(viper.GetString("server.log.output"))
				if err != nil {
					panic(fmt.Sprintf(
						"Error while creating log file [%s]: %s",
						viper.GetString("server.log.output"),
						err.Error(),
					))
				}
				defer f.Close()
			}
		}

		defaultLogger := middleware.DefaultLoggerConfig

		if viper.GetString("server.log.output") == "stdout" {
			log.SetOutput(os.Stdout)
			defaultLogger.Output = os.Stdout
		} else {
			f, _ := os.OpenFile(
				viper.GetString("server.log.output"),
				os.O_APPEND|os.O_CREATE|os.O_WRONLY,
				0775,
			)
			log.SetOutput(f)
			defaultLogger.Output = f
		}

		lvl := strings.ToLower(viper.GetString("server.log.level"))
		level, err := log.ParseLevel(lvl)

		if err != nil {
			level = log.InfoLevel
		}

		log.SetLevel(level)

		if viper.GetString("server.log.format") == "json" {
			log.SetFormatter(&log.JSONFormatter{})
		} else {
			log.SetFormatter(&log.TextFormatter{})
		}

		viper.SetDefault("config", config)

		e := echo.New()

		if viper.GetString("server.mode") == "dev" {
			e.Debug = true
		}

		e.Use(mid.Logger)
		e.Use(middleware.LoggerWithConfig(defaultLogger))
		e.Use(middleware.RequestID())
		e.Use(middleware.BodyLimit("2M"))
		e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Timeout: time.Duration(viper.GetInt("server.timeout")) * time.Second,
		}))
		e.Use(otelecho.Middleware(viper.GetString("server.name")))
		e.Use(middleware.Recover())

		p := prometheus.NewPrometheus(viper.GetString("server.name"), nil)
		p.Use(e)

		e.HTTPErrorHandler = func(err error, c echo.Context) {
			ctx := c.Request().Context()
			trace.SpanFromContext(ctx).RecordError(err)

			e.DefaultHTTPErrorHandler(err, c)
		}

		e.GET("/favicon.ico", func(c echo.Context) error {
			return c.String(http.StatusNoContent, "")
		})

		e.GET("/", controller.HealthAction)
		e.GET("/health", controller.HealthAction)

		var runerr error

		if viper.GetBool("server.tls.status") {
			runerr = e.StartTLS(
				fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("server.port"))),
				viper.GetString("server.tls.crt_path"),
				viper.GetString("server.tls.key_path"),
			)
		} else {
			runerr = e.Start(
				fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("server.port"))),
			)
		}

		if runerr != nil && runerr != http.ErrServerClosed {
			panic(runerr.Error())
		}
	},
}

func init() {
	serverCmd.Flags().StringVarP(
		&config,
		"config",
		"c",
		"config.prod.yml",
		"Absolute path to config file (required)",
	)
	serverCmd.MarkFlagRequired("config")
	rootCmd.AddCommand(serverCmd)
}
