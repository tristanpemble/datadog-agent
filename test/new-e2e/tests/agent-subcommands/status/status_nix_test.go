// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package status

import (
	"fmt"
	"testing"

	"github.com/DataDog/test-infra-definitions/components/datadog/agentparams"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	awshost "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/aws/host"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/utils/e2e/client"
)

type linuxStatusSuite struct {
	baseStatusSuite
}

func TestLinuxStatusSuite(t *testing.T) {
	t.Parallel()
	e2e.Run(t, &linuxStatusSuite{}, e2e.WithProvisioner(awshost.ProvisionerNoFakeIntake()))
}

func (v *linuxStatusSuite) TestStatusHostname() {
	metadata := client.NewEC2Metadata(v.T(), v.Env().RemoteHost.Host, v.Env().RemoteHost.OSFamily)
	resourceID := metadata.Get("instance-id")

	expectedSections := []expectedSection{
		{
			name:            `Hostname`,
			shouldBePresent: true,
			shouldContain:   []string{fmt.Sprintf("hostname: %v", resourceID), "hostname provider: aws"},
		},
	}

	fetchAndCheckStatus(&v.baseStatusSuite, expectedSections)
}

func (v *linuxStatusSuite) TestFIPSProxyStatus() {
	// Skip this test if the e2e pipeline is running with the FIPS Agent flavor because the FIPS proxy is not supported with the FIPS Agent.
	if v.Env().Agent.FIPSEnabled {
		v.T().Skip()
	}

	v.UpdateEnv(awshost.ProvisionerNoFakeIntake(awshost.WithAgentOptions(agentparams.WithAgentConfig("fips.enabled: true"))))

	expectedSections := []expectedSection{
		{
			name:            `Agent \(.*\)`,
			shouldBePresent: true,
			shouldContain:   []string{"FIPS Mode: proxy", "FIPS proxy"},
		},
	}

	fetchAndCheckStatus(&v.baseStatusSuite, expectedSections)
}

// This test asserts the presence of metadata sent by Python checks in the status subcommand output.
func (v *linuxStatusSuite) TestChecksMetadataUnix() {
	v.UpdateEnv(awshost.ProvisionerNoFakeIntake(awshost.WithAgentOptions(
		agentparams.WithFile("/etc/datadog-agent/conf.d/custom_check.yaml", string(customCheckYaml), true),
		agentparams.WithFile("/etc/datadog-agent/checks.d/custom_check.py", string(customCheckPython), true),
	)))

	expectedSections := []expectedSection{
		{
			name:            "Collector",
			shouldBePresent: true,
			shouldContain: []string{"Instance ID:", "[OK]",
				"metadata:",
				"custom_metadata_key: custom_metadata_value",
			},
		},
	}

	fetchAndCheckStatus(&v.baseStatusSuite, expectedSections)
}

func (v *linuxStatusSuite) TestDefaultInstallStatus() {
	v.testDefaultInstallStatus([]string{"Status: Not running or unreachable"}, nil)
}
