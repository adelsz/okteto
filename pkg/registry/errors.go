// Copyright 2022 The Okteto Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"fmt"
	"strings"

	oktetoErrors "github.com/okteto/okteto/pkg/errors"
)

// GetErrorMessage returns the parsed error message
func GetErrorMessage(err error, tag string) error {
	if err == nil {
		return nil
	}
	imageRegistry, imageTag := GetRegistryAndRepo(tag)
	switch {
	case IsLoggedIntoRegistryButDontHavePermissions(err):
		err = oktetoErrors.UserError{
			E:    fmt.Errorf("error building image '%s': You are not authorized to push image '%s'", tag, imageTag),
			Hint: fmt.Sprintf("Please log in into the registry '%s' with a user with push permissions to '%s' or use another image.", imageRegistry, imageTag),
		}
	case IsNotLoggedIntoRegistry(err):
		err = oktetoErrors.UserError{
			E:    fmt.Errorf("error building image '%s': You are not authorized to push image '%s'", tag, imageTag),
			Hint: fmt.Sprintf("Log in into the registry '%s' and verify that you have permissions to push the image '%s'.", imageRegistry, imageTag),
		}
	case IsBuildkitServiceUnavailable(err):
		err = oktetoErrors.UserError{
			E:    fmt.Errorf("buildkit service is not available at the moment"),
			Hint: "Please try again later.",
		}
	default:
		err = oktetoErrors.UserError{
			E: fmt.Errorf("error building image '%s': %s", tag, err.Error()),
		}
	}
	return err
}

// IsTransientError returns true if err represents a transient registry error
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	switch {
	case strings.Contains(err.Error(), "failed commit on ref") && strings.Contains(err.Error(), "500 Internal Server Error"),
		strings.Contains(err.Error(), "transport is closing"):
		return true
	case strings.Contains(err.Error(), "transport: error while dialing: dial tcp: i/o timeout"):
		return true
	case strings.Contains(err.Error(), "error reading from server: EOF"):
		return true
	case strings.Contains(err.Error(), "error while dialing: dial tcp: lookup buildkit") && strings.Contains(err.Error(), "no such host"):
		return true
	case strings.Contains(err.Error(), "failed commit on ref") && strings.Contains(err.Error(), "400 Bad Request"):
		return true
	case strings.Contains(err.Error(), "failed to do request") && strings.Contains(err.Error(), "http: server closed idle connection"):
		return true
	case strings.Contains(err.Error(), "failed to do request") && strings.Contains(err.Error(), "tls: use of closed connection"):
		return true
	case strings.Contains(err.Error(), "Canceled") && strings.Contains(err.Error(), "the client connection is closing"):
		return true
	case strings.Contains(err.Error(), "Canceled") && strings.Contains(err.Error(), "context canceled"):
		return true
	default:
		return false
	}
}

// IsLoggedIntoRegistryButDontHavePermissions returns true when the error is because the user is logged into the registry but doesn't have permissions to push the image
func IsLoggedIntoRegistryButDontHavePermissions(err error) bool {
	return strings.Contains(err.Error(), "insufficient_scope: authorization failed")
}

// IsNotLoggedIntoRegistry returns true when the error is because the user is not logged into the registry
func IsNotLoggedIntoRegistry(err error) bool {
	return strings.Contains(err.Error(), "failed to authorize: failed to fetch anonymous token") ||
		strings.Contains(err.Error(), "UNAUTHORIZED: authentication required")
}

// IsBuildkitServiceUnavailable returns true when an error is because buildkit is unavailable
func IsBuildkitServiceUnavailable(err error) bool {
	return strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), "500 Internal Server Error") || strings.Contains(err.Error(), "context canceled")
}
