/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/manisbindra/az-mpf/pkg/domain"
	"github.com/manisbindra/az-mpf/pkg/infrastructure/mpfSharedUtils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

var (
	defaultConfigFilename      = "stingoftheviper"
	envPrefix                  = "MPF"
	replaceHyphenWithCamelCase = false

	flgSubscriptionID     string
	flgTenantID           string
	flgSPClientID         string
	flgSPObjectID         string
	flgSPClientSecret     string
	flgShowDetailedOutput bool
	flgJSONOutput         bool
	flgVerbose            bool
	flgDebug              bool
	// RootCmd            *cobra.Command
)

func NewRootCommand() *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "az-mpf",
		Short: "Find minimum permissions required for Azure deployments",
		Long: `Find minimum permissions required for Azure deployments including ARM and Terraform based deployments. For example:
		
		This CLI allows you to find the minimum permissions required for Azure deployments including ARM and Terraform based deployments. 
		A Service Principal is required to run this CLI. All permissions associated with the Service principal are initially wiped by this command:`,
		Example: `az-mpf arm --subscriptionID <subscriptionID> --tenantID <tenantID> --spClientID <spClientID> --spObjectID <spObjectID> --spClientSecret <spClientSecret>
		az-mpm terraform --subscriptionID <subscriptionID> --tenantID <tenantID> --spClientID <spClientID> --spObjectID <spObjectID> --spClientSecret <spClientSecret> --executablePath <executablePath> --workingDir <workingDir> --varFilePath <varFilePath>
		`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	rootCmd.PersistentFlags().StringVarP(&flgSubscriptionID, "subscriptionID", "s", "", "Azure Subscription ID")
	rootCmd.PersistentFlags().StringVarP(&flgTenantID, "tenantID", "", "", "Azure Tenant ID")
	rootCmd.PersistentFlags().StringVarP(&flgSPClientID, "spClientID", "", "", "Service Principal Client ID")
	rootCmd.PersistentFlags().StringVarP(&flgSPObjectID, "spObjectID", "", "", "Service Principal Object ID")
	rootCmd.PersistentFlags().StringVarP(&flgSPClientSecret, "spClientSecret", "", "", "Service Principal Client Secret")
	rootCmd.PersistentFlags().BoolVarP(&flgShowDetailedOutput, "showDetailedOutput", "", false, "Show detailed output")
	rootCmd.PersistentFlags().BoolVarP(&flgJSONOutput, "jsonOutput", "", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&flgVerbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&flgDebug, "debug", "d", false, "debug output")

	rootCmd.MarkPersistentFlagRequired("subscriptionID")
	rootCmd.MarkPersistentFlagRequired("tenantID")
	rootCmd.MarkPersistentFlagRequired("spClientID")
	rootCmd.MarkPersistentFlagRequired("spObjectID")
	rootCmd.MarkPersistentFlagRequired("spClientSecret")

	rootCmd.MarkFlagsMutuallyExclusive("showDetailedOutput", "jsonOutput")

	// Add subcommands
	rootCmd.AddCommand(NewARMCommand())
	rootCmd.AddCommand(NewTerraformCommand())

	return rootCmd
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	v.SetConfigName(defaultConfigFilename)

	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	bindFlags(cmd, v)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name
		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		if replaceHyphenWithCamelCase {
			configName = strings.ReplaceAll(f.Name, "-", "")
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func setLogLevel() {
	if flgVerbose {
		log.SetLevel(log.InfoLevel)
	}
	if flgDebug {
		log.SetLevel(log.DebugLevel)
	}
}

func getRootMPFConfig() domain.MPFConfig {
	mpfRole := domain.Role{}

	roleDefUUID, _ := uuid.NewRandom()
	mpfRole.RoleDefinitionID = roleDefUUID.String()
	mpfRole.RoleDefinitionName = fmt.Sprintf("tmp-rol-%s", mpfSharedUtils.GenerateRandomString(7))
	mpfRole.RoleDefinitionResourceID = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s", flgSubscriptionID, mpfRole.RoleDefinitionID)
	log.Infoln("roleDefinitionResourceID:", mpfRole.RoleDefinitionResourceID)

	return domain.MPFConfig{
		SubscriptionID: flgSubscriptionID,
		TenantID:       flgTenantID,
		SP: domain.ServicePrincipal{
			SPClientID:     flgSPClientID,
			SPObjectID:     flgSPObjectID,
			SPClientSecret: flgSPClientSecret,
		},
		Role: mpfRole,
	}
}

func getAbsolutePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {

		absWorkingDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		absPath = absWorkingDir + "/" + absPath
	}
	return absPath, nil
}
