# Setting Default Service Accounts
You can configure default service accounts to use for specific GCP projects
using the `default-service-accounts` command. Service accounts configured using
this command will then be used as the default value in subsequent commands that
require the `--service-account-email` flag.

```
$ eiam default-service-accounts set
INFO    Using current project: my-project
INFO    Checking 123 service accounts in my-project
Use the arrow keys to navigate: ↓ ↑ → ←
Select Service Account
    svc-acct-1@my-project.iam.gserviceaccount.com
   ►  svc-acct-2@my-project.iam.gserviceaccount.com

INFO    Set default service account for my-project to svc-acct-2@my-project.iam.gserviceaccount.com
```

```
$ eiam gcloud compute instances list -R "Demonstrating default service accounts"

Reason ------------- ephemeral-iam 87fee575daeba1ae: Demonstrating default service accounts
Command ------------ gcloud compute instances list
Project ------------ my-project
Service Account ---- svc-acct-2@my-project.iam.gserviceaccount.com
```

You can set a default service account for multiple different projects:

```
$ eiam default-sa set --project another-project
INFO    Using current project: another-project
INFO    Checking 123 service accounts in another-project
Use the arrow keys to navigate: ↓ ↑ → ←
Select Service Account
   ►  different-svc-acct@another-project.iam.gserviceaccount.com
    different-svc-acct-2@another-project.iam.gserviceaccount.com

INFO    Set default service account for another-project to different-svc-acct@another-project.iam.gserviceaccount.com
```

```
$ eiam default-sa list

PROJECT            SERVICE ACCOUNT
my-project         svc-acct-2@my-project.iam.gserviceaccount.com
another-project    different-svc-acct@another-project.iam.gserviceaccount.com
```