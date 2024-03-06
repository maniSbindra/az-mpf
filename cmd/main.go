// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {

	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.ErrorLevel
	}
	log.SetLevel(logLevel)

	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}
