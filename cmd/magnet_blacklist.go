package cmd

import (
	"encoding/json"
	"errors"
	"os"

	_ "github.com/KeepShareOrg/keepshare/config" // Register the shared file/environment defaults without starting services.
	"github.com/KeepShareOrg/keepshare/hosts"
	pikpakapi "github.com/KeepShareOrg/keepshare/hosts/pikpak/api"
	pikpakquery "github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/magnetblock"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

func init() {
	command := &cobra.Command{
		Use:   "magnet-blacklist",
		Short: "Manage the platform-wide magnet blacklist (JSON output; no prompts)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return magnetBlacklistError(cmd, map[string]any{"action": "magnet-blacklist"}, "invalid_arguments", "Choose add, remove, or list")
		},
	}
	command.SetFlagErrorFunc(func(cmd *cobra.Command, _ error) error {
		return magnetBlacklistError(cmd, map[string]any{"action": cmd.Name()}, "invalid_arguments", "Invalid command flags; see --help")
	})
	for _, action := range []string{"add", "remove"} {
		action := action
		var hash string
		short := "Block a magnet globally, then revoke its existing PikPak shares; retries revoke again"
		if action == "remove" {
			short = "Remove only the global block; do not restore shares or change cloud files"
		}
		command.AddCommand(&cobra.Command{
			Use:   action + " <magnet>",
			Short: short,
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) != 1 {
					return magnetBlacklistError(cmd, map[string]any{"action": action}, "invalid_arguments", "Exactly one magnet URI is required")
				}
				var err error
				hash, err = magnetblock.Normalize(args[0])
				if err != nil {
					return magnetBlacklistError(cmd, map[string]any{"action": action}, "invalid_magnet", "Expected a valid magnet URI with one unambiguous BTIH info hash")
				}
				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return runMagnetBlacklist(cmd, action, hash, args[0])
			},
		})
	}
	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List global blocks ordered by creation time and info hash",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return magnetBlacklistError(cmd, map[string]any{"action": "list"}, "invalid_arguments", "List does not accept arguments")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMagnetBlacklist(cmd, "list", "", "")
		},
	})
	rootCmd.AddCommand(command)
	// The existing root prints returned errors through stdLog; keep this command's stdout JSON-only.
	if len(os.Args) > 1 && os.Args[1] == command.Name() {
		stdLog.SetOutput(os.Stderr)
		rootCmd.SilenceErrors = true
		rootCmd.SilenceUsage = true
	}
}

func magnetBlacklistError(cmd *cobra.Command, result map[string]any, code, message string) error {
	stdLog.SetOutput(os.Stderr)
	cmd.Root().SilenceErrors = true
	cmd.Root().SilenceUsage = true
	result["error"] = map[string]string{"code": code, "message": message}
	if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
		return errors.New("json_output_failed")
	}
	return errors.New(code)
}

func runMagnetBlacklist(cmd *cobra.Command, action, hash, link string) error {
	result := map[string]any{"action": action}
	if hash != "" {
		result["info_hash"] = hash
	}
	// Load the same configuration paths, but not config.Load's mail/client initialization.
	log.Log().SetOutput(os.Stderr)
	log.SetLevel("panic")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/keepshare/")
	viper.AddConfigPath("./conf/")
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return magnetBlacklistError(cmd, result, "config_load_failed", "Unable to read configuration")
		}
	}
	// A private, silent GORM logger prevents DSNs, queries, or driver errors from contaminating JSON or exposing credentials.
	db, err := gorm.Open(mysql.Open(viper.GetString("db_mysql")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return magnetBlacklistError(cmd, result, "mysql_open_failed", "Unable to connect to MySQL")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return magnetBlacklistError(cmd, result, "mysql_open_failed", "Unable to initialize MySQL")
	}
	defer sqlDB.Close()

	switch action {
	case "list":
		items := make([]model.MagnetBlacklist, 0)
		if err := db.WithContext(cmd.Context()).Clauses(dbresolver.Write).Order("created_at ASC").Order("info_hash ASC").Find(&items).Error; err != nil {
			return magnetBlacklistError(cmd, result, "blacklist_read_failed", "Unable to read the global blacklist")
		}
		result["items"] = items
	case "remove":
		deleted := db.WithContext(cmd.Context()).Clauses(dbresolver.Write).Where("info_hash = ?", hash).Delete(&model.MagnetBlacklist{})
		if deleted.Error != nil {
			return magnetBlacklistError(cmd, result, "blacklist_remove_failed", "Unable to remove the global block")
		}
		result["local_blocked"] = false
		result["removed"] = deleted.RowsAffected > 0
	case "add":
		inserted := db.WithContext(cmd.Context()).Clauses(dbresolver.Write, clause.OnConflict{DoNothing: true}).Create(&model.MagnetBlacklist{InfoHash: hash, OriginalLink: link})
		if inserted.Error != nil {
			return magnetBlacklistError(cmd, result, "blacklist_add_failed", "Unable to persist the global block; remote revocation was not attempted")
		}
		result["local_blocked"] = true
		result["already_blocked"] = inserted.RowsAffected == 0
		result["remote_attempted"] = false
		options, err := redis.ParseURL(viper.GetString("db_redis"))
		if err != nil {
			return magnetBlacklistError(cmd, result, "revocation_unavailable", "The global block is active, but Redis configuration is invalid; retry add after fixing configuration")
		}
		redisClient := redis.NewClient(options)
		defer redisClient.Close()
		maintenance := pikpakapi.NewMaintenance(pikpakquery.Use(db), &hosts.Dependencies{Mysql: db, Redis: redisClient})
		result["remote_attempted"] = true
		revoked, err := maintenance.RevokeMagnetShares(cmd.Context(), link)
		if revoked.Failures == nil {
			revoked.Failures = []pikpakapi.RevokeFailure{}
		}
		result["remote"] = revoked
		if err != nil || len(revoked.Failures) != 0 {
			return magnetBlacklistError(cmd, result, "revocation_incomplete", "The global block is active; remote revocation or local cleanup is incomplete; retry add to resume")
		}
	}
	if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
		return errors.New("json_output_failed")
	}
	return nil
}
