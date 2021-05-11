// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
)

// gRPC error protobuf type names.
const (
	RPCStatusBadRequest          = "type.googleapis.com/google.rpc.BadRequest"
	RPCStatusDebugInfo           = "type.googleapis.com/google.rpc.DebugInfo"
	RPCStatusErrorInfo           = "type.googleapis.com/google.rpc.ErrorInfo"
	RPCStatusHelp                = "type.googleapis.com/google.rpc.Help"
	RPCStatusPreconditionFailure = "type.googleapis.com/google.rpc.PreconditionFailure"
	RPCStatusQuotaFailure        = "type.googleapis.com/google.rpc.QuotaFailure"
	RPCStatusRetryInfo           = "type.googleapis.com/google.rpc.RetryInfo"
)

func checkGoogleRPCError(err error) EiamError {
	if serr, ok := err.(EiamError); ok {
		err = serr.Err
	}
	errDetails := map[string]string{}

	if serr, ok := status.FromError(err); ok {
		for _, detail := range serr.Proto().Details {
			switch detail.GetTypeUrl() {
			case RPCStatusBadRequest:
				errDetails["Bad Request"] = parseRPCStatusBadRequest(detail)
			case RPCStatusDebugInfo:
				errDetails["Debug Info"] = parseRPCStatusDebugInfo(detail)
			case RPCStatusErrorInfo:
				errDetails["Error Info"] = parseRPCStatusErrorInfo(detail)
			case RPCStatusHelp:
				errDetails["Help"] = parseRPCStatusHelp(detail)
			case RPCStatusPreconditionFailure:
				errDetails["Precondition Failure"] = parseRPCStatusPreconditionFailure(detail)
			case RPCStatusQuotaFailure:
				errDetails["Quota Failure"] = parseRPCStatusQuotaFailure(detail)
			case RPCStatusRetryInfo:
				errDetails["Retry Info"] = parseRPCStatusRetryInfo(detail)
			default:
				util.Logger.Debugf("Unrecognized gRPC status detail type: %s", detail.GetTypeUrl())
			}
		}
		errField := errors.New(serr.Message())
		if len(errDetails) > 0 {
			errMsg := "[gRPC Status Error]\n"
			for title, details := range errDetails {
				errMsg += fmt.Sprintf("[%s]\n%s\n", title, details)
			}
			return New(errMsg, errField).(EiamError)
		}
		return New("A gRPC error occurred. For more information, set the logging level to debug", errField).(EiamError)
	}
	return EiamError{}
}

func parseRPCStatusBadRequest(detail *anypb.Any) string {
	var buf strings.Builder
	badReq := &errdetails.BadRequest{}
	if err := detail.UnmarshalTo(badReq); err != nil {
		util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusBadRequest, err)
	} else if violations := badReq.GetFieldViolations(); len(violations) > 0 {
		fmt.Fprint(&buf, "  Violating Fields:\n")
		for _, viol := range badReq.GetFieldViolations() {
			fmt.Fprintf(&buf, "    %s: %s\n", viol.GetField(), viol.GetDescription())
		}
	}
	if buf.Len() > 0 {
		return buf.String()
	}
	return ""
}

func parseRPCStatusDebugInfo(detail *anypb.Any) string {
	var buf strings.Builder
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
	}
	if buf.Len() > 0 {
		return buf.String()
	}
	return ""
}

func parseRPCStatusErrorInfo(detail *anypb.Any) string {
	var buf strings.Builder
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
		// Example: {"resource": "projects/123", "service": "pubsub.googleapis.com"}.
		if metadata := errInfo.GetMetadata(); len(metadata) > 0 {
			fmt.Fprint(&buf, "  Metadata:\n")
			for k, v := range metadata {
				fmt.Fprintf(&buf, "    %s: %s\n", k, v)
			}
		}
	}
	if buf.Len() > 0 {
		return buf.String()
	}
	return ""
}

func parseRPCStatusHelp(detail *anypb.Any) string {
	var buf strings.Builder
	errHelp := &errdetails.Help{}
	if err := detail.UnmarshalTo(errHelp); err != nil {
		util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusHelp, err)
	} else if links := errHelp.GetLinks(); len(links) > 0 {
		// URL(s) pointing to additional information on handling the current error.
		for _, link := range links {
			fmt.Fprintf(&buf, "  %s: %s\n", link.GetDescription(), link.GetUrl())
		}
	}
	if buf.Len() > 0 {
		return buf.String()
	}
	return ""
}

func parseRPCStatusPreconditionFailure(detail *anypb.Any) string {
	var buf strings.Builder
	preconFailure := &errdetails.PreconditionFailure{}
	if err := detail.UnmarshalTo(preconFailure); err != nil {
		util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusPreconditionFailure, err)
	} else if violations := preconFailure.GetViolations(); len(violations) > 0 {
		fmt.Fprint(&buf, "  Violations Details:\n")
		for _, viol := range violations {
			fmt.Fprintf(&buf, "    [%s] %s: %s\n", viol.GetType(), viol.GetSubject(), viol.GetDescription())
		}
	}
	if buf.Len() > 0 {
		return buf.String()
	}
	return ""
}

func parseRPCStatusQuotaFailure(detail *anypb.Any) string {
	var buf strings.Builder
	quotaFailure := &errdetails.QuotaFailure{}
	if err := detail.UnmarshalTo(quotaFailure); err != nil {
		util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusQuotaFailure, err)
	} else if violations := quotaFailure.GetViolations(); len(violations) > 0 {
		fmt.Fprint(&buf, "  Violations Details:\n")
		for _, viol := range violations {
			fmt.Fprintf(&buf, "    %s: %s\n", viol.GetSubject(), viol.GetDescription())
		}
	}
	if buf.Len() > 0 {
		return buf.String()
	}
	return ""
}

func parseRPCStatusRetryInfo(detail *anypb.Any) string {
	var buf strings.Builder
	retryInfo := &errdetails.RetryInfo{}
	if err := detail.UnmarshalTo(retryInfo); err != nil {
		util.Logger.Errorf("Failed to unmarshal %s from Any protobuf: %v", RPCStatusRetryInfo, err)
	} else {
		delay := retryInfo.RetryDelay.Seconds
		fmt.Fprintf(&buf, "Please wait %d seconds before retrying request\n", delay)
	}
	if buf.Len() > 0 {
		return buf.String()
	}
	return ""
}
