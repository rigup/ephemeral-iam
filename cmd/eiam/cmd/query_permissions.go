package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd/options"
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
	queryiam "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/gcpclient/query_iam"
)

// Resource string templates
var (
	computeInstanceRes = "//compute.googleapis.com/projects/%s/zones/%s/instances/%s"
	projectsRes        = "//cloudresourcemanager.googleapis.com/projects/%s"
	pubsubTopicsRes    = "//pubsub.googleapis.com/projects/%s/topics/%s"
	serviceAccountsRes = "//iam.googleapis.com/projects/%s/serviceAccounts/%s"
	storageBucketsRes  = "//storage.googleapis.com/projects/_/buckets/%s"
)

var queryPermsCmdConfig options.CmdConfig

func newCmdQueryPermissions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-permissions",
		Short: "Query current permissions on a GCP resource",
		Long: dedent.Dedent(`
			Compare the set of permissions you've been granted against a full list of possible permissions on a resource.
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
			testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resourceString)
			if err != nil {
				return err
			}
			userPerms, err := queryiam.QueryComputeInstancePermissions(
				testablePerms,
				queryPermsCmdConfig.Project,
				queryPermsCmdConfig.Zone,
				queryPermsCmdConfig.ComputeInstance,
			)
			if err != nil {
				return err
			}
			printPermissions(uniq(testablePerms), userPerms)
			return nil
		},
	}

	options.AddProjectFlag(cmd.Flags(), &queryPermsCmdConfig.Project)
	options.AddZoneFlag(cmd.Flags(), &queryPermsCmdConfig.Zone, true)
	options.AddComputeInstanceFlag(cmd.Flags(), &queryPermsCmdConfig.ComputeInstance, true)

	return cmd
}

func newCmdQueryProjectPermissions() *cobra.Command {
	var resourceString string
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Query the permissions you are granted at the project level",
		PreRun: func(cmd *cobra.Command, args []string) {
			resourceString = fmt.Sprintf(projectsRes, queryPermsCmdConfig.Project)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resourceString)
			if err != nil {
				return err
			}
			userPerms, err := queryiam.QueryProjectPermissions(testablePerms, queryPermsCmdConfig.Project)
			if err != nil {
				return err
			}
			printPermissions(uniq(testablePerms), userPerms)
			return nil
		},
	}

	options.AddProjectFlag(cmd.Flags(), &queryPermsCmdConfig.Project)

	return cmd
}

func newCmdQueryPubSubPermissions() *cobra.Command {
	var resourceString string
	cmd := &cobra.Command{
		Use:   "pubsub",
		Short: "Query the permissions you are granted on a pubsub topic",
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)
			resourceString = fmt.Sprintf(pubsubTopicsRes, queryPermsCmdConfig.Project, queryPermsCmdConfig.PubSubTopic)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resourceString)
			if err != nil {
				return err
			}
			userPerms, err := queryiam.QueryPubSubPermissions(testablePerms, queryPermsCmdConfig.Project, queryPermsCmdConfig.PubSubTopic)
			if err != nil {
				return err
			}
			printPermissions(uniq(testablePerms), userPerms)
			return nil
		},
	}

	options.AddProjectFlag(cmd.Flags(), &queryPermsCmdConfig.Project)
	options.AddPubSubTopicFlag(cmd.Flags(), &queryPermsCmdConfig.PubSubTopic, true)

	return cmd
}

func newCmdQueryServiceAccountPermissions() *cobra.Command {
	var resourceString string
	cmd := &cobra.Command{
		Use:   "service-account",
		Short: "Query the permissions you are granted on a service account",
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)
			resourceString = fmt.Sprintf(serviceAccountsRes, queryPermsCmdConfig.Project, queryPermsCmdConfig.ServiceAccountEmail)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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
			printPermissions(uniq(testablePerms), userPerms)
			return nil
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
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)
			resourceString = fmt.Sprintf(storageBucketsRes, queryPermsCmdConfig.StorageBucket)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			testablePerms, err := queryiam.QueryTestablePermissionsOnResource(resourceString)
			if err != nil {
				return err
			}
			userPerms, err := queryiam.QueryStorageBucketPermissions(testablePerms, queryPermsCmdConfig.StorageBucket)
			if err != nil {
				return err
			}
			printPermissions(uniq(testablePerms), userPerms)
			return nil
		},
	}

	options.AddStorageBucketFlag(cmd.Flags(), &queryPermsCmdConfig.StorageBucket, true)

	return cmd
}

func printPermissions(fullPerms, userPerms []string) {
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)

	if len(userPerms) == 0 {
		fmt.Println("AVAILABLE")
		for _, perm := range fullPerms {
			fmt.Println(perm)
		}
		w.Flush()
		fmt.Println()
		util.Logger.Warn("You do not have any access to this resource")
	} else if len(userPerms) == len(fullPerms) {
		fmt.Fprintln(w, "AVAILABLE\tGRANTED")
		for i, perm := range fullPerms {
			fmt.Fprintf(w, "%s\t%s\n", perm, userPerms[i])
		}
		w.Flush()
		fmt.Println()
		util.Logger.Info("You have full access to this resource")
	} else {
		diff := difference(fullPerms, userPerms)

		// Pad slices to make even column lengths
		for i := range fullPerms {
			if i >= len(userPerms) {
				userPerms = append(userPerms, " ")
			}
			if i >= len(diff) {
				diff = append(diff, " ")
			}
		}

		fmt.Fprintln(w, "AVAILABLE\tGRANTED\tMISSING")
		for i, perm := range fullPerms {
			fmt.Fprintf(w, "%s\t%s\t%s\n", perm, userPerms[i], diff[i])
		}
		w.Flush()
	}
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func uniq(a []string) []string {
	mb := make(map[string]struct{}, len(a))
	for _, x := range a {
		mb[x] = struct{}{}
	}
	set := make([]string, 0, len(mb))
	for k := range mb {
		set = append(set, k)
	}
	sort.Strings(set)
	return set
}
