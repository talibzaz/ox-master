package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use: "ox",
	Short: "Eventackle ticketing",
	Long: "Eventackle ticketing microservice",
}