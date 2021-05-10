package eiam

import (
	"fmt"
	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

var tokenName string

func NewCmdPluginsAuth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Github authentication for installing plugins from private repos",
		Long: `
To use the "eiam plugins install" command with private Github repos, ephemeral-iam
needs to make authenticated calls to the Github API using a Github personal access
token. The "eiam plugins auth" subcommands allow you to configure ephemeral-iam to
make authenticated calls to Github and manage the credentials that it uses.

You can configure multiple tokens at once and designate which token to use by
referencing the name that it was given when it was added (defaults to "default").
		`,
	}
	cmd.AddCommand(NewCmdPluginsAuthAdd())
	cmd.AddCommand(NewCmdPluginsAuthList())
	cmd.AddCommand(NewCmdPluginsAuthDelete())
	return cmd
}

func NewCmdPluginsAuthAdd() *cobra.Command {
	tokenConfig := viper.GetStringMapString(appconfig.GithubTokens)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a Github personal access token to use to install plugins from private repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := tokenConfig[tokenName]; ok {
				util.Logger.Warnf("A token with the name %s already exists", tokenName)
				prompt := promptui.Prompt{
					Label:     fmt.Sprintf("Overwrite %s", tokenName),
					IsConfirm: true,
				}

				if _, err := prompt.Run(); err != nil {
					util.Logger.Warnf("Abandoning deletion of token: %s", tokenName)
					return nil
				}
			}
			util.Logger.Infof("Adding token with the name %s", tokenName)
			prompt := promptui.Prompt{
				Label: "Enter your Github Personal Access Token: ",
				Mask:  '●',
				Validate: func(input string) error {
					if len(input) != 40 {
						return fmt.Errorf("incorrect input length: expected 40, got %d", len(input))
					}
					tokenRegex := regexp.MustCompile(`^(?:ghp_)?[[:alnum:]]{36}(?:[[:alnum:]]{4})?$`)
					if !tokenRegex.MatchString(input) {
						return fmt.Errorf("invalid input: input was not a valid personal access token")
					}
					return nil
				},
			}

			token, err := prompt.Run()
			if err != nil {
				return errorsutil.New("User input failed", err)
			}

			tokenConfig[tokenName] = token
			viper.Set(appconfig.GithubTokens, tokenConfig)
			if !viper.GetBool(appconfig.GithubAuth) {
				viper.Set(appconfig.GithubAuth, true)
			}

			if err := viper.WriteConfig(); err != nil {
				return errorsutil.New("Failed to write updated configuration", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&tokenName, "name", "n", "default", "The name associated with the target access token")
	return cmd
}

func NewCmdPluginsAuthList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the existing Github personal access tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !viper.GetBool(appconfig.GithubAuth) {
				util.Logger.Warn("No Github credentials are currently configured.")
				return nil
			}
			fmt.Println("\n GITHUB ACCESS TOKENS\n----------------------")
			for tokenName := range viper.GetStringMapString(appconfig.GithubTokens) {
				fmt.Println(tokenName)
			}
			fmt.Printf("\n")
			return nil
		},
	}
	return cmd
}

func NewCmdPluginsAuthDelete() *cobra.Command {
	tokenConfig := viper.GetStringMapString(appconfig.GithubTokens)
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing Github personal access token from the config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenName == "" {
				tn, err := util.SelectToken(tokenConfig)
				if err != nil {
					return errorsutil.New("Failed to select access token", err)
				}
				tokenName = tn
			}
			if _, ok := tokenConfig[tokenName]; !ok {
				err := fmt.Errorf("no token with the name %s exists in the config", tokenName)
				return errorsutil.New("Failed to delete token from config", err)
			}

			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Delete %s", tokenName),
				IsConfirm: true,
			}

			if _, err := prompt.Run(); err != nil {
				util.Logger.Warnf("Abandoning deletion of token: %s", tokenName)
				return nil
			}

			delete(tokenConfig, tokenName)
			viper.Set(appconfig.GithubTokens, tokenConfig)

			if len(tokenConfig) == 0 {
				viper.Set(appconfig.GithubAuth, false)
			}

			if err := viper.WriteConfig(); err != nil {
				return errorsutil.New("Failed to write updated configuration", err)
			}
			util.Logger.Infof("%s was successfully deleted", tokenName)

			return nil
		},
	}
	cmd.Flags().StringVarP(&tokenName, "name", "n", "", "The name associated with the target access token")
	return cmd
}
