/*
Copyright 2026 The KubeVela Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bcode

// ErrAIConsoleNotConfigured is returned when no AI console endpoint is set,
// so there is nothing to open.
var ErrAIConsoleNotConfigured = NewBcode(503, 25001, "the AI console is not connected; set --ai-console-endpoint on the VelaUX server")

// ErrAIConsoleProjectRequired is returned for a workspace ticket asked
// without a project to scope it to.
var ErrAIConsoleProjectRequired = NewBcode(400, 25002, "pick a project to open the AI workspace in")
