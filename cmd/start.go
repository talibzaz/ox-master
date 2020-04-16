package cmd

import (
	"github.com/spf13/cobra"
	"github.com/graphicweave/ox/http"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"time"
	"github.com/spf13/viper"
	"github.com/graphicweave/ox/grpc"
	log "github.com/sirupsen/logrus"
)

func init() {
	RootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:  "start",
	Long: "start OX",
	PreRun: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.JSONFormatter{})
		viper.AutomaticEnv()
	},
	Run: func(cmd *cobra.Command, args []string) {

		var gracefulStop = make(chan os.Signal)

		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		signal.Notify(gracefulStop, syscall.SIGKILL)

		go func() {
			sig := <-gracefulStop
			fmt.Printf("caught sig: %+v\n", sig)
			log.Info("Wait for 2 second to finish processing")
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}()

		log.Info("Started gRPC server on :9001")
		go func() {
			if err := grpc.StartServer(viper.GetString("GRPC_ADDR")); err != nil {
				log.Error("error while starting ox gRPC server")
				log.Fatal(err.Error())
			}
		}()

		log.Infoln("started server on :9090")
		if err := http.StartServer(viper.GetString("HTTP_HOST")); err != nil {
			log.Info("error while starting ox HTTP server")
			log.Fatal(err.Error())
		}
	},
}
