package migrator

import (
	"github.com/spf13/cobra"
	"fmt"
	"github.com/golang-migrate/migrate"
	"os"
	"errors"
	"strings"
)

const clusterFlag = "cluster"
const versionFlag = "version"

// Initialize the Migrate with driver based on cluster
type MigrationClientFunc func(cluster string) (*migrate.Migrate, error)

var mcf MigrationClientFunc

var migrationClient *migrate.Migrate
var currentVersion uint

var rootCmd = &cobra.Command{
	Use:   "migration-client",
	Short: "migration client",
	Long: `A custom migration client built on top of 
			https://github.com/golang-migrate/migrate`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var dirty bool
		// fetch cluster of migration being executed
		cluster, err := cmd.PersistentFlags().GetString(clusterFlag)
		if err != nil {
			return err
		} else if strings.TrimSpace(cluster) == "" {
			return errors.New("cluster is missing")
		}

		// get the initialized migrationClient based on cluster
		migrationClient, err = mcf(strings.TrimSpace(cluster))

		// print current version before execution
		currentVersion, dirty, err = migrationClient.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return err
		}
		fmt.Printf("\nBefore execution: Version %d dirty %t \n", currentVersion, dirty)
		return nil
	},
}

var upCommand = &cobra.Command{
	Use:   "up",
	Short: "Apply all change or by version",
	Long:  `Applies all migration upwards`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := cmd.Flags().GetUint(versionFlag)
		if err != nil {
			fmt.Println("Invalid version value: ", err)
		}
		if version > 0 {
			// apply migration until given version - where input version is greater than current version
			if version <= currentVersion {
				return errors.New("version should be greater than current version")
			}
			err = migrationClient.Migrate(version)
		} else {
			// apply all remaining migrations
			err = migrationClient.Up()
		}

		if err != nil && err != migrate.ErrNilVersion && err != migrate.ErrNoChange {
			return err
		}
		return nil
	},
}

var forceCommand = &cobra.Command{
	Use:   "force",
	Short: "forcefully mark fake migration for given version",
	Long:  `Marks version as applied irrespective of it being dirty`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := cmd.Flags().GetUint(versionFlag)
		if err != nil {
			return err
		}

		if version > 0 {
			// force apply any version
			err = migrationClient.Force(int(version))
		}

		if err != nil && err != migrate.ErrNilVersion && err != migrate.ErrNoChange {
			return err
		}
		return nil
	},
}

var downCommand = &cobra.Command{
	Use:   "down",
	Short: "Undo migration to given version",
	Long:  `Applies down migration of all migration until given version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := cmd.Flags().GetUint(versionFlag)
		if err != nil {
			return err
		}
		if version < 1 || version >= currentVersion {
			// apply down migration until given version - where input version is lesser than current version
			return errors.New("version should be lesser than current version")
		}
		err = migrationClient.Migrate(version)

		if err != nil && err != migrate.ErrNilVersion && err != migrate.ErrNoChange {
			return err
		}
		return nil
	},
}

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Fetch current version",
	Long:  `Display current version and status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, dirty, err := migrationClient.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return err
		}
		fmt.Printf("\nversion: Current Version %d dirty %t \n", version, dirty)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringP(clusterFlag, "c", "", "cluster on which migration will be performed")
	rootCmd.MarkPersistentFlagRequired(clusterFlag)
	rootCmd.PersistentFlags().SortFlags = false

	//up command - version is optional
	upCommand.Flags().UintP(versionFlag, "v", 0, "optional: version of migration to be applied")

	//force command
	forceCommand.Flags().UintP(versionFlag, "v", 0, "version of migration to be applied")
	forceCommand.MarkFlagRequired(versionFlag)

	//down command
	downCommand.Flags().UintP(versionFlag, "v", 0, "version of migration to be applied")
	downCommand.MarkFlagRequired(versionFlag)

	rootCmd.AddCommand(upCommand)
	rootCmd.AddCommand(forceCommand)
	rootCmd.AddCommand(downCommand)
	rootCmd.AddCommand(versionCommand)
}

func cleanUp() {
	if migrationClient != nil {
		//fmt.Println("Exiting after cleanup")
		migrationClient.Close()
	}
}

func Execute(migrationClientFunc MigrationClientFunc) {
	mcf = migrationClientFunc
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error executing command - ", r)
			cleanUp()
			os.Exit(1)
		}
	}()
	err := rootCmd.Execute()
	cleanUp()
	if err != nil {
		fmt.Println("Error executing command - ", err)
		os.Exit(1)
	}
}
