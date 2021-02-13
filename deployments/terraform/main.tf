
resource "random_string" "suffix" {
  upper   = "false"
  lower   = "true"
  special = "false"
  length  = 6
}

locals {
  random_id = random_string.suffix.result
}

provider "google" {
  project = var.project
  region  = var.location
}

data "google_project" "project" {
  project_id = var.project
}

resource "google_project_service" "services" {
  project = var.project
  for_each = toset([
    "iam.googleapis.com",
    "iamcredentials.googleapis.com",
    "logging.googleapis.com",
    "pubsub.googleapis.com",
  ])
  service            = each.value
  disable_on_destroy = false
}

#############################################
#                  Pub/Sub                  #
#############################################

resource "google_pubsub_topic" "log_export" {
  project = data.google_project.project.project_id
  name    = "ephemeral-iam-audit-logs-topic-${local.random_id}"
}

resource "google_pubsub_subscription" "log_export" {
  project = data.google_project.project.project_id
  name    = "ephemeral-iam-audit-logs-sumo-${local.random_id}"
  topic   = google_pubsub_topic.log_export.id

  push_config {
    push_endpoint = var.sumo_http_source_url
  }
}

resource "google_logging_project_sink" "log_export" {
  project     = data.google_project.project.project_id
  name        = "ephemeral-iam-audit-log-sink-${local.random_id}"
  destination = "pubsub.googleapis.com/projects/${var.project}/topics/${google_pubsub_topic.log_export.name}"

  filter      = var.log_export_filter

  unique_writer_identity = true
}

resource "google_pubsub_topic_iam_member" "publisher" {
  project = data.google_project.project.project_id
  topic   = google_pubsub_topic.log_export.id
  role    = "roles/pubsub.publisher"
  member  = google_logging_project_sink.log_export.writer_identity
}
