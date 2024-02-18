package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/krugdenis/kubeconfig-generator/pkg"
)

var email string
var clusterServer string
var clusterRole string
var skipIP bool
var deleteSa bool

var rootCmd = &cobra.Command{
	Use:   "your_app_name",
	Short: "Your short description",
	Long:  `Your long description`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.Execute(email, clusterServer, clusterRole, skipIP, deleteSa)
	},
}

func init() {
	rootCmd.Flags().StringVar(&email, "email", "", "Email address (required)")
	rootCmd.Flags().StringVar(&clusterServer, "ip", "", "Cluster IP address (optional)")
	rootCmd.Flags().StringVar(&clusterRole, "cr", "", "Custom cluster role yaml (optional)")
	rootCmd.Flags().BoolVar(&skipIP, "skipIP", false, "Skip providing IP and use the default from the selected context")
	rootCmd.Flags().BoolVar(&deleteSa, "delete", false, "Skip creating service account, delete only")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
