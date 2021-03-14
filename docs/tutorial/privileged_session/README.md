# Using a Privileged Session

The core feature of the `eiam` CLI is its ability to grant users a short-lived privileged session.
This is accomplished using the `assume-privileges` (or just `priv`) command. When you run this command, `eiam`
will drop you into a shell that is configured to use the permissions of the provided service account. Detailed information
about how the session is handled can be found in the **Conceptual Overview** section of the [README](../../../README.md).

> **It is not currently feasible to test the sub-shell for complete POSIX compliance; therefore, if you come across any bugs**
> **in its implementation, please add the details of said bug to the issue linked above.**

## Example Workflow
Say a user (UserA) needs to debug an issue in a Cloud PubSub Topic. To debug the issue, UserA needs to publish test messages
to the topic, but their requests are being denied due to insufficient privileges.  UserA's workflow
could look something like this:

1. Query their permissions on the PubSub topic
2. List the service accounts that they have access to assume the privileges of
3. Query the service account's permissions on the PubSub topic
4. Confirm that the service account has access by running a one-off `gcloud` command using `eiam`
5. Start a privileged session using the `assume-privileges` command to publish their test messages

### Query their permissions on the PubSub topic
First, UserA may want to debug why they are unable to publish messages to the topic using the `query-permissions` command:

```
$ eiam query-permissions pubsub --topic example-topic

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
```

As the command output shows, UserA only has the ability to get PubSub topics, but they do not have the `pubsub.topics.publish`
permission that they need.

### List the available service accounts
Next, UserA can check and see if any of the service accounts that they have been given access to has the ability to
publish messages to the topic.  First, UserA gets a list of service accounts that they can access using the `list-service-accounts` command:

```
INFO    Using current project: example-project

EMAIL                                                   DESCRIPTION
pubsub-admin@example-project.iam.gserviceaccount.com    Service account that grants admin access on Cloud Pub/Sub topics
```

### Query the permissions granted to the service account
Now that UserA has a service account email that they can use, they can ensure that it has the `pubsub.topics.publish` permission
that they need to debug the topic:

```
$ eiam query-permissions pubsub --topic example-topic -s pubsub-admin@example-project.iam.gserviceaccount.com

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

INFO    pubsub-admin@example-project.iam.gserviceaccount.com has full access to this resource
```

Great!  This service account has the permission UserA needs to debug the topic.

### Confirm access to topic
Before starting a privileged session, it might be worth testing that the service account can indeed publish messages
to the topic so UserA uses the `eiam gcloud` command:

```
$ eiam gcloud pubsub topics publish projects/example-project/topics/example-topic --message="Testing" \
  --service-account-email pubsub-admin@example-project.iam.gserviceaccount.com \
  --reason "Debugging Pub/Sub topic (JIRA-1234)"

Project ------------ example-project
Service Account ---- pubsub-admin@example-project.iam.gserviceaccount.com
Reason ------------- Debugging Pub/Sub topic (JIRA-1234)
Command ------------ gcloud pubsub topics publish projects/example-project/topics/example-topic --message=Testing

Continue: y
INFO    Running: [gcloud pubsub topics publish projects/example-project/topics/example-topic --message=Testing]

messageIds:
- '2124890400294542'
```

### Starting a privileged debugging session
Now UserA can start a short-lived privileged session as the service account to continue debugging the Pub/Sub topic:

```
$ eiam assume-privileges \
  --service-account-email pubsub-admin@example-project.iam.gserviceaccount.com \
  --reason "Debugging Pub/Sub topic (JIRA-1234)"

Project ------------ example-project
Service Account ---- pubsub-admin@example-project.iam.gserviceaccount.com
Reason ------------- Debugging Pub/Sub topic (JIRA-1234)

Continue: y
INFO    Fetching short-lived access token for pubsub-admin@example-project.iam.gserviceaccount.com
INFO    Configuring gcloud to use auth proxy
INFO    Starting auth proxy
INFO    Privileged session will last until Tue, 09 Mar 2021 09:08:33 CST
WARNING Press CTRL+C to quit privileged session

[pubsub-admin@example-project.iam.gserviceaccount.com]
[eiam] > gcloud pubsub topics publish projects/example-project/topics/example-topic --message="Testing"
messageIds:
- '2125113463491038'

[pubsub-admin@example-project.iam.gserviceaccount.com]
[eiam] > 
```

This privileged session will last for 10 minutes and `eiam` will exit either when that time is up, or when
UserA closes the sub-shell using `CTRL-C`.

## Using `kubectl`
There is currently an [open bug](https://github.com/jessesomerville/ephemeral-iam/issues/2) regarding the behavior
of `kubectl` when using the `assume-permissions` command. In a future version of Ephemeral IAM, `kubectl` commands
will automatically be authenticated as the service account that is being impersonated, but for now, you must authenticate
to a GKE cluster after starting the sub-shell.  This will generate a new entry in your `.kubeconfig` which can then be used
to make calls to the cluster as the service account.

**Start a privileged session:**
```
$ eiam assume-privileges \
  --service-account-email gke-debug@example-project.iam.gserviceaccount.com \
  --reason "Debugging GKE workload (JIRA-1234)"
```

**Authenticate to the GKE cluster:**
```
[gke-debug@example-project.iam.gserviceaccount.com]
[eiam] > gcloud container clusters get-credentials break-glass-test --zone us-central1-c
Fetching cluster endpoint and auth data.
kubeconfig entry generated for break-glass-test.

[gke-debug@example-project.iam.gserviceaccount.com]
[eiam] > kubectl get pods
NAME                            READY   STATUS    RESTARTS   AGE
redis-master-6b54579d85-7swfn   1/1     Running   0          5d16h
```