# Running a Single Command
There are some use-cases where a user only needs to run a single `gcloud` or `kubectl` command with privileged
access.  For convenience purposes, `eiam` provides the ability to run one-off `gcloud` and `kubectl` commands.
The output from the `gcloud` and `kubectl` commands are sent to stdout to support redirection to pipes.

## Running a gcloud command

```
$ eiam gcloud compute instances list \
  --service-account-email compute-debug@example-project.iam.gserviceaccount.com \
  --reason "JIRA-1234"

Reason ------------- JIRA-1234
Command ------------ gcloud compute instances list
Project ------------ example-project
Service Account ---- compute-debug@example-project.iam.gserviceaccount.com

Continue: y
INFO    Running: [gcloud compute instances list]

NAME                                             ZONE           MACHINE_TYPE  PREEMPTIBLE  INTERNAL_IP    EXTERNAL_IP    STATUS
ephemeral-iam-demo                               us-central1-a  e2-medium                  10.128.15.193  35.223.80.157  RUNNING
gke-break-glass-test-default-pool-f489f36f-arxt  us-central1-c  e2-medium                  10.128.15.195  35.223.226.30  RUNNING
gke-break-glass-test-default-pool-f489f36f-kx0d  us-central1-c  e2-medium                  10.128.15.194  34.68.232.126  RUNNING
gke-break-glass-test-default-pool-f489f36f-tiu3  us-central1-c  e2-medium                  10.128.15.196  35.232.218.37  RUNNING
```

## Running a kubectl command
One off commands can also be used for long running tasks such as port-forwarding to a deployment in a GKE cluster:

```
$ eiam kubectl port-forward deployment/redis-master 7000:6379
  --service-account-email gke-debug@example-project.iam.gserviceaccount.com \
  --reason "JIRA-1234"

Reason ------------- JIRA-1234
Command ------------ kubectl port-forward deployment/redis-master 7000:6379
Project ------------ example-project
Service Account ---- compute-debug@example-project.iam.gserviceaccount.com

Continue: y
INFO    Fetching access token for gke-debug@example-project.iam.gserviceaccount.com
INFO    Running: [kubectl port-forward deployment/redis-master 7000:6379]
```

## Running cloud_sql_proxy

```
$ eiam cloud_sql_proxy -instances my-project:us-central1:example-instance=tcp:3306 \
	--service-account-email example@my-project.iam.gserviceaccount.com \
	--reason "Debugging for (JIRA-1234)"

Command ------------ cloud_sql_proxy -instances my-project:us-central1:example-instance=tcp:3306
Project ------------ my-project
Service Account ---- example@my-project.iam.gserviceaccount.com
Reason ------------- ephemeral-iam 968be336d4b769e2: Debugging for (JIRA-1234)

Continue: y
INFO    Fetching access token for example@my-project.iam.gserviceaccount.com
INFO    Running: [cloud_sql_proxy -instances my-project:us-central1:example-instance=tcp:3306]

2021/04/29 03:24:17 current FDs rlimit set to 1048576, wanted limit is 8500. Nothing to do here.
2021/04/29 03:24:18 Listening on 127.0.0.1:3306 for my-project:us-central1:example-instance
2021/04/29 03:24:18 Ready for new connections
```