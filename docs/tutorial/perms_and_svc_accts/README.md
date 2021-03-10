# Inspecting Permissions and Service Accounts

`eiam` allows you to debug permission issues by querying the permissions granted to both you and any service
accounts that you have access to impersonate.

## List Available Service Accounts

You can view which service accounts you have access to assume the privileges of using the `list-service-accounts` command.

```
$ eiam list-service-accounts

EMAIL                                         DESCRIPTION
svc-acct-1@project.iam.gserviceaccount.com    Privileged access to connect to SQL databases
svc-acct-2@project.iam.gserviceaccount.com    Editor access in the project
```

## Debugging Permissions

You can debug issues with permissions using the `query-permissions` command.  This command allows you to
check which permissions have been granted on a given resource.  As of `eiam` v0.0.4, only a few resources are supported
as the process is different for each resource type.  The resources that are currently supported are:

- Compute Instances
- Project Level Permissions
- Cloud PubSub
- Service Accounts
- Storage Buckets

Permissions can be queried for both your default user account and any service accounts that you have access to impersonate.

More information about the output format of the `query-permissions` command is located in the command's `help`:

```
$ eiam query-permissions --help

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

Usage:
  eiam query-permissions [command]

Available Commands:
  compute-instance Query the permissions you are granted on a compute instance
  project          Query the permissions you are granted at the project level
  pubsub           Query the permissions you are granted on a pubsub topic
  service-account  Query the permissions you are granted on a service account
  storage-bucket   Query the permissions you are granted on a storage bucket

Flags:
  -h, --help   help for query-permissions

Global Flags:
  -y, --yes   Assume 'yes' to all prompts

Use "eiam query-permissions [command] --help" for more information about a command.
```

> **For brevity's sake, outputs have been redacted from the commands shown below.**

### Query Permissions Granted on Compute Instances

```
$ eiam query-permissions compute-instance \
  --zone us-central1-a --instance my-instance

$ eiam query-permissions compute-instance \
  --zone us-central1-a --instance my-instance \
  --service-account-email example@my-project.iam.gserviceaccount.com
```

### Query Permissions Granted at the Project Level

```
$ eiam query-permissions project
```

### Query Permissions Granted on a PubSub Topic

```
$ eiam query-permissions pubsub -t topic1

$ eiam query-permissions pubsub -t topic1 \
  --service-account-email example@my-project.iam.gserviceaccount.com
```

### Query Permissions Granted on a Service Account

```
$ eiam query-permissions service-account \
  --service-account-email example@my-project.iam.gserviceaccount.com
```

### Query Permissions Granted on a Storage Bucket

```
$ eiam query-permissions storage-bucket --bucket bucket-name

$ eiam query-permissions storage-bucket --bucket bucket-name \
  --service-account-email example@my-project.iam.gserviceaccount.com
```