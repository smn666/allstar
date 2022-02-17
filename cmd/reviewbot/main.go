// Copyright 2022 Allstar Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"os"

	"github.com/ossf/allstar/pkg/webhandler"
	"github.com/rs/zerolog"
)

func main() {
	setupLog()

	// Use go 'flags' module
	if len(os.Args) < 2 {
		log.Fatal("Need an arg")
	}

	// get from disk, future: get from cloud secrets

	os.Environ()

	privateKeyName := os.Args[1]

	webhandler.HandleWebhooks(privateKeyName)
}

func setupLog() {
	// Match expected values in GCP
	zerolog.LevelFieldName = "severity"
	zerolog.LevelTraceValue = "DEFAULT"
	zerolog.LevelDebugValue = "DEBUG"
	zerolog.LevelInfoValue = "INFO"
	zerolog.LevelWarnValue = "WARNING"
	zerolog.LevelErrorValue = "ERROR"
	zerolog.LevelFatalValue = "CRITICAL"
	zerolog.LevelPanicValue = "CRITICAL"
}
