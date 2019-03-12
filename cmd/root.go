package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"

	jiralib "github.com/andygrunwald/go-jira"
	gh "github.com/google/go-github/v24/github"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"

	"github.com/ystia/zenhub-jira-sync/pkg"
	"github.com/ystia/zenhub-jira-sync/pkg/clients/github"
	"github.com/ystia/zenhub-jira-sync/pkg/clients/jira"
	"github.com/ystia/zenhub-jira-sync/pkg/clients/zenhub"
)

var rootCmd = &cobra.Command{
	Use:          "zenhub-jira-sync",
	Short:        "Synchronize ZenHub/GitHub issues to JIRA",
	SilenceUsage: true,
	RunE: func(c *cobra.Command, args []string) error {
		cfg := new(Config)

		viper.Unmarshal(cfg)

		err := validateConfig(cfg)
		if err != nil {
			return err
		}
		for _, s := range cfg.Synchronizations {
			err := syncRepository(cfg, s)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file (default is /etc/zh-jira-sync/zh-jira-sync.[json|yaml])")

}

var cfgFile string

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cobra" (without extension).

		//Configuration file directories
		viper.SetConfigName("zh-jira-sync") // name of config file (without extension)
		viper.AddConfigPath("/etc/zh-jira-sync/")
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
}

func createGithubClient(ctx context.Context, cfg *Config) *gh.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubAPIToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return gh.NewClient(tc)
}

func createJiraClient(cfg *Config) (*jiralib.Client, error) {
	tp := jiralib.BasicAuthTransport{
		Username: cfg.JiraAuthentication.User,
		Password: cfg.JiraAuthentication.Password,
	}
	jiraClient, err := jiralib.NewClient(tp.Client(), cfg.JiraURI)
	return jiraClient, errors.Wrapf(err, "failed to create jira client")
}

func syncRepository(cfg *Config, s Synchronization) error {
	ctx := context.Background()
	jiraClient, err := createJiraClient(cfg)
	if err != nil {
		return err
	}

	syncJiraClient := &jira.Client{
		JiraClient: jiraClient,
		BoardID:    s.JiraBoardID,
		ProjectKey: cfg.JiraProjectKey,
	}
	err = syncJiraClient.Init()
	if err != nil {
		return err
	}

	sync := &pkg.Sync{
		GithubClient: &github.Client{
			GHClient: createGithubClient(ctx, cfg),
			Owner:    s.GithubOwner,
			Repo:     s.GithubRepository,
		},
		JiraClient: syncJiraClient,
	}

	if s.IssueLabelToType == nil {
		// If not found a synchronization level look at global level
		s.IssueLabelToType = cfg.IssueLabelToType
		if s.IssueLabelToType == nil {
			// If no mapping use a default
			s.IssueLabelToType = &IssueLabelToType{
				Default: "User story",
			}
		}
	}

	sync.DefaultIssueType = s.IssueLabelToType.Default
	sync.LabelsToIssueType = make([]struct {
		Label     string
		IssueType string
	}, 0)
	for _, ltt := range s.IssueLabelToType.LabelsMapping {
		for label, issueType := range ltt {
			sync.LabelsToIssueType = append(sync.LabelsToIssueType, struct {
				Label     string
				IssueType string
			}{
				label,
				issueType,
			})
		}
	}

	ghRepo, err := sync.GithubClient.GetRepository(ctx)
	if err != nil {
		return err
	}
	sync.ZenhubClient = zenhub.NewClient(cfg.ZenhubAPIToken, *ghRepo.ID)

	if s.ReleaseRenamer.Source == "" {
		s.ReleaseRenamer.Source = "^(.*)$"
	}

	if s.ReleaseRenamer.Target == "" {
		s.ReleaseRenamer.Target = "${0}"
	}
	sync.ReleaseNameRE, err = regexp.Compile(s.ReleaseRenamer.Source)
	if err != nil {
		return errors.Wrapf(err, "failed to compile release regexp for repository %s/%s", s.GithubOwner, s.GithubRepository)
	}
	sync.VersionNameRename = s.ReleaseRenamer.Target

	sync.DefaultJiraComponents = getSyncJiraComponents(cfg.DefaultJiraComponents, s.DefaultJiraComponents)

	return sync.All(ctx)
}

func getSyncJiraComponents(globalComponents []string, repositoryComponents []string) []string {
	// Note(loicalbertin): this is not very efficient but it's an easy way to remove duplicates
	dedupMap := make(map[string]struct{}, len(globalComponents)+len(repositoryComponents))
	repositoryComponents = append(repositoryComponents, globalComponents...)
	for _, comp := range repositoryComponents {
		dedupMap[comp] = struct{}{}
	}
	repositoryComponents = repositoryComponents[:0]
	for comp := range dedupMap {
		repositoryComponents = append(repositoryComponents, comp)
	}
	return repositoryComponents
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
