// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package run

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DataDog/datadog-agent/cmd/agent/command"
	"github.com/DataDog/datadog-agent/comp/core"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
)

func TestCommand(t *testing.T) {
	fxutil.TestOneShotSubcommand(t,
		Commands(&command.GlobalParams{}),
		[]string{"run"},
		run,
		func(cliParams *cliParams, coreParams core.BundleParams) {
			require.Equal(t, true, coreParams.ConfigLoadSecrets)
		})
}

func TestCommandPidfile(t *testing.T) {
	fxutil.TestOneShotSubcommand(t,
		Commands(&command.GlobalParams{}),
		[]string{"run", "--pidfile", "/pid/file"},
		run,
		func(cliParams *cliParams, coreParams core.BundleParams) {
			require.Equal(t, "/pid/file", cliParams.pidfilePath)
			require.Equal(t, true, coreParams.ConfigLoadSecrets)
		})
}
