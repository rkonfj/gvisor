// Copyright 2023 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sync

// TraceBlockReason constants, from Go's src/runtime/trace.go.
const (
	TraceBlockSelect TraceBlockReason = traceEvGoBlockSelect // +checkconst runtime traceBlockSelect
	TraceBlockSync                    = traceEvGoBlockSync   // +checkconst runtime traceBlockSync
)

// Tracer event types, from Go's src/runtime/trace.go.
const (
	traceEvGoBlockSelect = 24 // +checkconst runtime traceEvGoBlockSelect
	traceEvGoBlockSync   = 25 // +checkconst runtime traceEvGoBlockSync
)
