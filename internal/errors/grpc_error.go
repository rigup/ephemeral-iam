package errors

import (
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"

	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
)

const (
	RPCStatusBadRequest          = "type.googleapis.com/google.rpc.BadRequest"
	RPCStatusDebugInfo           = "type.googleapis.com/google.rpc.DebugInfo"
	RPCStatusErrorInfo           = "type.googleapis.com/google.rpc.ErrorInfo"
	RPCStatusHelp                = "type.googleapis.com/google.rpc.Help"
	RPCStatusPreconditionFailure = "type.googleapis.com/google.rpc.PreconditionFailure"
	RPCStatusQuotaFailure        = "type.googleapis.com/google.rpc.QuotaFailure"
	RPCStatusRetryInfo           = "type.googleapis.com/google.rpc.RetryInfo"
)

func checkGoogleRPCError(serr *EiamError) *EiamError {
	err := serr.Err
	errDetails := map[string]string{}

	if serr, ok := status.FromError(err); ok {
		for _, detail := range serr.Proto().Details {
			var buf strings.Builder
			switch detail.GetTypeUrl() {
			case RPCStatusBadRequest:
				badReq := &errdetails.BadRequest{}
				if err := detail.UnmarshalTo(badReq); err != nil {
					util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusBadRequest, err)
				} else if violations := badReq.GetFieldViolations(); len(violations) > 0 {
					fmt.Fprint(&buf, "  Violating Fields:\n")
					for _, viol := range badReq.GetFieldViolations() {
						fmt.Fprintf(&buf, "    %s: %s\n", viol.GetField(), viol.GetDescription())
					}
					errDetails["Bad Request"] = buf.String()
				}
			case RPCStatusDebugInfo:
				debugInfo := &errdetails.DebugInfo{}
				if err := detail.UnmarshalTo(debugInfo); err != nil {
					util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusDebugInfo, err)
				} else {
					traces := debugInfo.GetStackEntries()
					details := debugInfo.GetDetail()
					if len(traces) > 0 {
						fmt.Fprintf(&buf, "  Stack Trace:\n    %s", strings.Join(traces, "\n    "))
					}
					if len(details) > 0 {
						fmt.Fprintf(&buf, "  Details:\n    %s", details)
					}
					if buf.Len() > 0 {
						errDetails["Debug Info"] = buf.String()
					}
				}
			case RPCStatusErrorInfo:
				errInfo := &errdetails.ErrorInfo{}
				if err := detail.UnmarshalTo(errInfo); err != nil {
					util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusErrorInfo, err)
				} else {
					domain := errInfo.GetDomain()
					reason := errInfo.GetReason()
					if len(domain) > 0 && len(reason) > 0 {
						fmt.Fprintf(&buf, "  Reason:\n    %s: %s\n", domain, reason)
					} else if len(domain) > 0 {
						fmt.Fprintf(&buf, "  Domain:\n    %s\n", domain)
					} else if len(reason) > 0 {
						fmt.Fprintf(&buf, "  Reason:\n    %s\n", reason)
					}

					// Additional structured details about this error.
					// Example: {"resource": "projects/123", "service": "pubsub.googleapis.com"}
					if metadata := errInfo.GetMetadata(); len(metadata) > 0 {
						fmt.Fprint(&buf, "  Metadata:\n")
						for k, v := range metadata {
							fmt.Fprintf(&buf, "    %s: %s\n", k, v)
						}
					}

					if buf.Len() > 0 {
						errDetails["Error Info"] = buf.String()
					}
				}
			case RPCStatusHelp:
				errHelp := &errdetails.Help{}
				if err := detail.UnmarshalTo(errHelp); err != nil {
					util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusHelp, err)
				} else if links := errHelp.GetLinks(); len(links) > 0 {
					// URL(s) pointing to additional information on handling the current error.
					for _, link := range links {
						fmt.Fprintf(&buf, "  %s: %s\n", link.GetDescription(), link.GetUrl())
					}
					errDetails["Help"] = buf.String()
				}
			case RPCStatusPreconditionFailure:
				preconFailure := &errdetails.PreconditionFailure{}
				if err := detail.UnmarshalTo(preconFailure); err != nil {
					util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusPreconditionFailure, err)
				} else if violations := preconFailure.GetViolations(); len(violations) > 0 {
					fmt.Fprint(&buf, "  Violations Details:\n")
					for _, viol := range violations {
						fmt.Fprintf(&buf, "    [%s] %s: %s\n", viol.GetType(), viol.GetSubject(), viol.GetDescription())
					}
					errDetails["Precondition Failure"] = buf.String()
				}
			case RPCStatusQuotaFailure:
				quotaFailure := &errdetails.QuotaFailure{}
				if err := detail.UnmarshalTo(quotaFailure); err != nil {
					util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusQuotaFailure, err)
				} else if violations := quotaFailure.GetViolations(); len(violations) > 0 {
					fmt.Fprint(&buf, "  Violations Details:\n")
					for _, viol := range violations {
						fmt.Fprintf(&buf, "    %s: %s\n", viol.GetSubject(), viol.GetDescription())
					}
					errDetails["Quota Failure"] = buf.String()
				}
			case RPCStatusRetryInfo:
				retryInfo := &errdetails.RetryInfo{}
				if err := detail.UnmarshalTo(retryInfo); err != nil {
					util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusRetryInfo, err)
				} else {
					errDetails["Retry Info"] = fmt.Sprintf("Please wait %d seconds before retrying request\n", retryInfo.RetryDelay.Seconds)
				}
			default:
				util.Logger.Debugf("Unrecognized gRPC status detail type: %s", detail.GetTypeUrl())
			}
		}

		if len(errDetails) > 0 {
			errMsg := "[gRPC Status Error]\n"
			for title, details := range errDetails {
				errMsg += fmt.Sprintf("[%s]\n%s\n", title, details)
			}
			return &EiamError{
				Log: util.Logger.WithError(err),
				Msg: errMsg,
				Err: err,
			}
		} else {
			return &EiamError{
				Log: util.Logger.WithError(err),
				Msg: fmt.Sprintf("[gRPC Status Error]: %v", serr.Message()),
				Err: err,
			}
		}
	}
	return nil
}
