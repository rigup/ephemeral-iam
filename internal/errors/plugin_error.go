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
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
)

var PkgVersionErr = regexp.MustCompile(`^plugin\.Open\("(.*)"\).*different version of package (.*)$`)

// CheckPluginError checks an error generated while loading plugins. If the error
// is an issue with the plugin, the error is logged, the plugin is not loaded,
// nil is returned, and execution continues.  If it is an unhandled error, it is
// returned and execution is halted.
func CheckPluginError(err error) error {
	errStr := err.Error()
	errFields := logrus.Fields{}
	var plugin string
	if res := PkgVersionErr.FindAllStringSubmatch(errStr, -1); res != nil {
		// There is a discrepancy between package versions in the plugin and eiam.
		groups := res[0]
		pluginPath := strings.Split(groups[1], "/")
		plugin = pluginPath[len(pluginPath)-1]
		pkg := groups[2]
		if strings.Contains(pkg, "ephemeral-iam") {
			errFields["error"] = "plugin was built with a different version of eiam"
		} else {
			errFields["error"] = fmt.Sprintf("plugin was built with a different version of %s", pkg)
		}
	}
	if len(errFields) == 0 || plugin == "" {
		return EiamError{
			Log: util.Logger.WithError(err),
			Msg: "Failed to load plugin",
			Err: err,
		}
	}
	util.Logger.WithFields(errFields).Errorf("Failed to load plugin %s", plugin)
	return nil
}
