variable "project" {
  description = "The GCP project ID"
  type        = string
}

variable "location" {
  description = "The location (region or zone) to host the cluster in"
  type        = string
}

variable "sumo_http_source_url" {
  description = "The Sumo Logic Source endpoint to export audit logs to"
  type        = string
}

variable "log_export_filter" {
  description = "The logging query used to filter audit logs to export"
  type        = string

  default = <<EOD
protoPayload.methodName="GenerateAccessToken"
NOT protoPayload.requestMetadata.requestAttributes.reason:*
EOD
}