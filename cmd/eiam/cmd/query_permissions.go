package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd/options"
	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
	"github.com/jessesomerville/ephemeral-iam/internal/gcpclient"
	queryiam "github.com/jessesomerville/ephemeral-iam/internal/gcpclient/query_iam"
)

// Resource string templates
var (
	computeInstanceRes = "//compute.googleapis.com/projects/%s/zones/%s/instances/%s"
	projectsRes        = "//cloudresourcemanager.googleapis.com/projects/%s"
	pubsubTopicsRes    = "//pubsub.googleapis.com/projects/%s/topics/%s"
	serviceAccountsRes = "//iam.googleapis.com/projects/%s/serviceAccounts/%s"
	storageBucketsRes  = "//storage.googleapis.com/projects/_/buckets/%s"

	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()
)

var queryPermsCmdConfig options.CmdConfig

func newCmdQueryPermissions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-permissions",
		Short: "Query current permissions on a GCP resource",
		Long: dedent.Dedent(`
			Compare the list of permissions granted on a resource against the full list of
			grantable permissions.
			
			For example, the list of grantable permissions on a Cloud PubSub Topic are as follows:
			
				pubsub.topics.attachSubscription
				pubsub.topics.delete
				pubsub.topics.detachSubscription
				pubsub.topics.get
				pubsub.topics.getIamPolicy
				pubsub.topics.publish
				pubsub.topics.setIamPolicy
				pubsub.topics.update
				pubsub.topics.updateTag
			
			Say a user (user1) is granted the PubSub Viewer role on a topic (topic1). The PubSub Viewer role grants the 
			"pubsub.topics.get" permissions on Topics.
			
				$ eiam query-permissions pubsub -t topic1
			
				AVAILABLE                           GRANTED
				pubsub.topics.attachSubscription    ✖
				pubsub.topics.delete                ✖
				pubsub.topics.detachSubscription    ✖
				pubsub.topics.get                   ✔
				pubsub.topics.getIamPolicy          ✖
				pubsub.topics.publish               ✖
				pubsub.topics.setIamPolicy          ✖
				pubsub.topics.update                ✖
				pubsub.topics.updateTag             ✖
			
			If user1 can assume the privileges of a service account (sa1), they can query the permissions that sa1
			has on the topic. Say sa1 has been granted the PubSub Admin role on topic1:
			
				$ eiam query-permissions pubsub -t topic1 -s sa1@project.iam.gserviceaccount.com
			
				AVAILABLE                           GRANTED
				pubsub.topics.attachSubscription    ✔
				pubsub.topics.delete                ✔
				pubsub.topics.detachSubscription    ✔
				pubsub.topics.get                   ✔
				pubsub.topics.getIamPolicy          ✔
				pubsub.topics.publish               ✔
				pubsub.topics.setIamPolicy          ✔
				pubsub.topics.update                ✔
				pubsub.topics.updateTag             ✔
			
				INFO    sa1@project.iam.gserviceaccount.com has full access to this resource
		`),
	}

	cmd.AddCommand(newCmdQueryComputeInstancePermissions())
	cmd.AddCommand(newCmdQueryProjectPermissions())
	cmd.AddCommand(newCmdQueryPubSubPermissions())
	cmd.AddCommand(newCmdQueryServiceAccountPermissions())
	cmd.AddCommand(newCmdQueryStorageBucketPermissions())

	return cmd
}

func newCmdQueryComputeInstancePermissions() *cobra.Command {
	var resourceString string
	cmd := &cobra.Command{
		Use:   "compute-instance",
		Short: "Query the permissions you are granted on a compute instance",
		Example: dedent.Dedent(`
			  eiam query-permissions compute-instance \
			    --zone us-central1-a --instance my-instance
			
			  eiam query-permissions compute-instance \
			    --zone us-central1-a --instance my-instance \
			    --service-account-email example@my-project.iam.gserviceaccount.com
		`),
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)
			resourceString = fmt.Sprintf(
				computeInstanceRes,
				queryPermsCmdConfig.Project,
				queryPermsCmdConfig.Zone,
				queryPermsCmdConfig.ComputeInstance,
			)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			util.Logger.Infof("Querying permissions granted on %s", resourceString)
			testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resourceString)
			if err != nil {
				util.Logger.Warnf("gcloud is configured to use %s as the default zone. If this is not correct, please provide the zone using the `--zone` flag", queryPermsCmdConfig.Zone)
				return err
			}
			userPerms, err := queryiam.QueryComputeInstancePermissions(
				testablePerms,
				queryPermsCmdConfig.Project,
				queryPermsCmdConfig.Zone,
				queryPermsCmdConfig.ComputeInstance,
				queryPermsCmdConfig.ServiceAccountEmail,
				queryPermsCmdConfig.Reason,
			)
			if err != nil {
				return err
			}
			if svcAcct := queryPermsCmdConfig.ServiceAccountEmail; svcAcct != "" {
				return printPermissions(util.Uniq(testablePerms), userPerms, svcAcct)
			} else {
				userAcct, err := gcpclient.CheckActiveAccountSet()
				if err != nil {
					return err
				}
				return printPermissions(util.Uniq(testablePerms), userPerms, userAcct)
			}
		},
	}

	options.AddProjectFlag(cmd.Flags(), &queryPermsCmdConfig.Project)
	options.AddZoneFlag(cmd.Flags(), &queryPermsCmdConfig.Zone, true)
	options.AddComputeInstanceFlag(cmd.Flags(), &queryPermsCmdConfig.ComputeInstance, true)
	options.AddServiceAccountEmailFlag(cmd.Flags(), &queryPermsCmdConfig.ServiceAccountEmail, false)
	options.AddReasonFlag(cmd.Flags(), &queryPermsCmdConfig.Reason, false)

	return cmd
}

func newCmdQueryProjectPermissions() *cobra.Command {
	var resourceString string
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Query the permissions you are granted at the project level",
		Example: "  eiam query-permissions project",
		PreRun: func(cmd *cobra.Command, args []string) {
			resourceString = fmt.Sprintf(projectsRes, queryPermsCmdConfig.Project)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			util.Logger.Infof("Querying permissions granted on %s", resourceString)
			testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resourceString)
			if err != nil {
				return err
			}
			userPerms, err := queryiam.QueryProjectPermissions(
				testablePerms,
				queryPermsCmdConfig.Project,
				queryPermsCmdConfig.ServiceAccountEmail,
				queryPermsCmdConfig.Reason,
			)
			if err != nil {
				return err
			}
			if svcAcct := queryPermsCmdConfig.ServiceAccountEmail; svcAcct != "" {
				return printPermissions(util.Uniq(testablePerms), userPerms, svcAcct)
			} else {
				userAcct, err := gcpclient.CheckActiveAccountSet()
				if err != nil {
					return err
				}
				return printPermissions(util.Uniq(testablePerms), userPerms, userAcct)
			}
		},
	}

	options.AddProjectFlag(cmd.Flags(), &queryPermsCmdConfig.Project)
	options.AddServiceAccountEmailFlag(cmd.Flags(), &queryPermsCmdConfig.ServiceAccountEmail, false)
	options.AddReasonFlag(cmd.Flags(), &queryPermsCmdConfig.Reason, false)

	return cmd
}

func newCmdQueryPubSubPermissions() *cobra.Command {
	var resourceString string
	cmd := &cobra.Command{
		Use:   "pubsub",
		Short: "Query the permissions you are granted on a pubsub topic",
		Example: dedent.Dedent(`
			  eiam query-permissions pubsub -t topic1
				
			  eiam query-permissions pubsub -t topic1 \
			    --service-account-email example@my-project.iam.gserviceaccount.com
		`),
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)
			resourceString = fmt.Sprintf(pubsubTopicsRes, queryPermsCmdConfig.Project, queryPermsCmdConfig.PubSubTopic)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			util.Logger.Infof("Querying permissions granted on %s", resourceString)
			testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resourceString)
			if err != nil {
				return err
			}
			userPerms, err := queryiam.QueryPubSubPermissions(
				testablePerms,
				queryPermsCmdConfig.Project,
				queryPermsCmdConfig.PubSubTopic,
				queryPermsCmdConfig.ServiceAccountEmail,
				queryPermsCmdConfig.Reason,
			)
			if err != nil {
				return err
			}
			if svcAcct := queryPermsCmdConfig.ServiceAccountEmail; svcAcct != "" {
				return printPermissions(util.Uniq(testablePerms), userPerms, svcAcct)
			} else {
				userAcct, err := gcpclient.CheckActiveAccountSet()
				if err != nil {
					return err
				}
				return printPermissions(util.Uniq(testablePerms), userPerms, userAcct)
			}
		},
	}

	options.AddProjectFlag(cmd.Flags(), &queryPermsCmdConfig.Project)
	options.AddPubSubTopicFlag(cmd.Flags(), &queryPermsCmdConfig.PubSubTopic, true)
	options.AddServiceAccountEmailFlag(cmd.Flags(), &queryPermsCmdConfig.ServiceAccountEmail, false)
	options.AddReasonFlag(cmd.Flags(), &queryPermsCmdConfig.Reason, false)

	return cmd
}

func newCmdQueryServiceAccountPermissions() *cobra.Command {
	var resourceString string
	cmd := &cobra.Command{
		Use:   "service-account",
		Short: "Query the permissions you are granted on a service account",
		Example: dedent.Dedent(`
			  eiam query-permissions service-account \
			    --service-account-email example@my-project.iam.gserviceaccount.com
		`),
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)
			resourceString = fmt.Sprintf(serviceAccountsRes, queryPermsCmdConfig.Project, queryPermsCmdConfig.ServiceAccountEmail)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			util.Logger.Infof("Querying permissions granted on %s", resourceString)
			testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resourceString)
			if err != nil {
				return err
			}
			userPerms, err := queryiam.QueryServiceAccountPermissions(
				testablePerms,
				queryPermsCmdConfig.Project,
				queryPermsCmdConfig.ServiceAccountEmail,
			)
			if err != nil {
				return err
			}
			if svcAcct := queryPermsCmdConfig.ServiceAccountEmail; svcAcct != "" {
				return printPermissions(util.Uniq(testablePerms), userPerms, svcAcct)
			} else {
				userAcct, err := gcpclient.CheckActiveAccountSet()
				if err != nil {
					return err
				}
				return printPermissions(util.Uniq(testablePerms), userPerms, userAcct)
			}
		},
	}

	options.AddServiceAccountEmailFlag(cmd.Flags(), &queryPermsCmdConfig.ServiceAccountEmail, true)
	options.AddProjectFlag(cmd.Flags(), &queryPermsCmdConfig.Project)

	return cmd
}

func newCmdQueryStorageBucketPermissions() *cobra.Command {
	var resourceString string
	cmd := &cobra.Command{
		Use:   "storage-bucket",
		Short: "Query the permissions you are granted on a storage bucket",
		Example: dedent.Dedent(`
			  eiam query-permissions storage-bucket --bucket bucket-name
			
			  eiam query-permissions storage-bucket --bucket bucket-name \
			    --service-account-email example@my-project.iam.gserviceaccount.com
		`),
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)
			resourceString = fmt.Sprintf(storageBucketsRes, queryPermsCmdConfig.StorageBucket)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			util.Logger.Infof("Querying permissions granted on %s", resourceString)
			testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resourceString)
			if err != nil {
				return err
			}
			userPerms, err := queryiam.QueryStorageBucketPermissions(
				testablePerms,
				queryPermsCmdConfig.StorageBucket,
				queryPermsCmdConfig.ServiceAccountEmail,
				queryPermsCmdConfig.Reason,
			)
			if err != nil {
				return err
			}
			if svcAcct := queryPermsCmdConfig.ServiceAccountEmail; svcAcct != "" {
				return printPermissions(util.Uniq(testablePerms), userPerms, svcAcct)
			} else {
				userAcct, err := gcpclient.CheckActiveAccountSet()
				if err != nil {
					return err
				}
				return printPermissions(util.Uniq(testablePerms), userPerms, userAcct)
			}
		},
	}

	options.AddStorageBucketFlag(cmd.Flags(), &queryPermsCmdConfig.StorageBucket, true)
	options.AddServiceAccountEmailFlag(cmd.Flags(), &queryPermsCmdConfig.ServiceAccountEmail, false)
	options.AddReasonFlag(cmd.Flags(), &queryPermsCmdConfig.Reason, false)

	return cmd
}

func printPermissions(fullPerms, userPerms []string, acctEmail string) error {
	if len(fullPerms) > 100 {
		// If the list of permissions is really long and the user has the less command
		// available, pipe the command to less to paginate the output
		lessPath, err := appconfig.CheckCommandExists("less")
		if err != nil {
			printPermissionsList(os.Stderr, fullPerms, userPerms, acctEmail, true)
		}

		// Create command for less with a stdin pipe that we can write to
		cmd := exec.Command(lessPath)
		cmd.Stdout = os.Stdout
		stdin, err := cmd.StdinPipe()
		if err != nil {
			util.Logger.Error("Failed to create stdin pipe for less command")
			return err
		}

		// Write the output in a goroutine so less can be ready to read it
		go func() {
			defer stdin.Close()
			printPermissionsList(stdin, fullPerms, userPerms, acctEmail, false)
		}()
		if err := cmd.Run(); err != nil {
			printPermissionsList(os.Stderr, fullPerms, userPerms, acctEmail, true)
		}
	} else {
		printPermissionsList(os.Stderr, fullPerms, userPerms, acctEmail, true)
	}
	return nil
}

func printPermissionsList(out io.Writer, fullPerms, userPerms []string, acctEmail string, color bool) {
	yes, no := "✔", "✖"
	if color {
		yes, no = green(yes), red(no)
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 4, 4, ' ', 0)

	fmt.Fprintln(w, "AVAILABLE\tGRANTED")

	for _, perm := range fullPerms {
		if util.Contains(userPerms, perm) {
			fmt.Fprintf(w, "%s\t%s\n", perm, yes)
		} else {
			fmt.Fprintf(w, "%s\t%s\n", perm, no)
		}
	}
	w.Flush()
	fmt.Fprintf(out, "\n%s\n\n", buf.String())
	fmt.Println()

	if len(userPerms) == 0 {
		util.Logger.Warnf("%s does not have any access to this resource", acctEmail)
	} else if len(userPerms) == len(fullPerms) {
		util.Logger.Infof("%s has full access to this resource", acctEmail)
	}
}
