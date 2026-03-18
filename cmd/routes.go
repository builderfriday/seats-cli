package cmd

import (
	"fmt"
	"os"

	"github.com/derek/seats-cli/internal/api"
	"github.com/derek/seats-cli/internal/format"
	"github.com/spf13/cobra"
)

var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "List routes for a program",
	RunE:  runRoutes,
}

func init() {
	routesCmd.Flags().String("program", "", "Mileage program")
	_ = routesCmd.MarkFlagRequired("program")
	rootCmd.AddCommand(routesCmd)
}

func runRoutes(cmd *cobra.Command, args []string) error {
	program, _ := cmd.Flags().GetString("program")

	client := api.NewClient(os.Getenv("SEATS_AERO_API_KEY"))
	routes, err := client.Routes(program)
	if err != nil {
		return handleAPIError(err)
	}

	fmt.Print(format.FormatRoutesMarkdown(program, routes))
	return nil
}
