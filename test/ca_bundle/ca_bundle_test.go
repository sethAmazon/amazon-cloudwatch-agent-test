// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

//go:build !windows

package ca_bundle

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/amazon-cloudwatch-agent-test/environment"
	"github.com/aws/amazon-cloudwatch-agent-test/internal/common"
)

const (
	configOutputPath       = "/opt/aws/amazon-cloudwatch-agent/bin/config.json"
	commonConfigOutputPath = "/opt/aws/amazon-cloudwatch-agent/etc/common-config.toml"
	configJSON             = "/config.json"
	commonConfigTOML       = "/common-config.toml"
	targetString           = "x509: certificate signed by unknown authority"

	// Let the agent run for 30 seconds. This will give agent enough time to call server
	agentRuntime         = 30 * time.Second
	localstackS3Key      = "integration-test/ls_tmp/%s"
	keyDelimiter         = "/"
	localstackConfigPath = "../../localstack/ls_tmp/"
	originalPem          = "original.pem"
	combinePem           = "combine.pem"
	snakeOilPem          = "snakeoil.pem"
	tmpDirectory         = "/tmp/"
)

type input struct {
	findTarget bool
	dataInput  string
}

var envMetaDataStrings = &(environment.MetaDataStrings{})

func init() {
	environment.RegisterEnvironmentMetaDataFlags(envMetaDataStrings)
}

// Must run this test with parallel 1 since this will fail if more than one test is running at the same time
// This test uses a pem file created for the local stack endpoint to be able to connect via ssl
func TestBundle(t *testing.T) {
	metadata := environment.GetEnvironmentMetaData(envMetaDataStrings)
	t.Logf("metadata required for test cwa sha %s bucket %s ca cert path %s", metadata.CwaCommitSha, metadata.Bucket, metadata.CaCertPath)
	setUpLocalstackConfig(metadata)

	parameters := []input{
		//Use the system pem ca bundle  + local stack pem file ssl should connect thus target string not found
		{dataInput: "resources/integration/ssl/with/combine/bundle", findTarget: false},
		//Do not look for ca bundle with http connection should connect thus target string not found
		{dataInput: "resources/integration/ssl/without/bundle/http", findTarget: false},
		//Use the system pem ca bundle ssl should not connect thus target string found
		{dataInput: "resources/integration/ssl/with/original/bundle", findTarget: true},
		//Do not look for ca bundle should not connect thus target string found
		{dataInput: "resources/integration/ssl/without/bundle", findTarget: true},
	}

	for _, parameter := range parameters {
		//before test run
		log.Printf("resource file location %s find target %t", parameter.dataInput, parameter.findTarget)
		t.Run(fmt.Sprintf("resource file location %s find target %t", parameter.dataInput, parameter.findTarget), func(t *testing.T) {
			common.ReplaceLocalStackHostName(parameter.dataInput + configJSON)
			t.Logf("config file after localstack host replace %s", string(readFile(parameter.dataInput+configJSON)))
			common.CopyFile(parameter.dataInput+configJSON, configOutputPath)
			common.CopyFile(parameter.dataInput+commonConfigTOML, commonConfigOutputPath)
			common.StartAgent(configOutputPath, true)
			time.Sleep(agentRuntime)
			log.Printf("Agent has been running for : %s", agentRuntime.String())
			common.StopAgent()
			output := common.ReadAgentOutput(agentRuntime)
			containsTarget := outputLogContainsTarget(output)
			if (parameter.findTarget && !containsTarget) || (!parameter.findTarget && containsTarget) {
				t.Errorf("Find target is %t contains target is %t", parameter.findTarget, containsTarget)
			}
		})
	}
}

func outputLogContainsTarget(output string) bool {
	log.Printf("Log file %s", output)
	contains := strings.Contains(output, targetString)
	log.Printf("Log file contains target string %t", contains)
	return contains
}

// Get localstack pem files
func setUpLocalstackConfig(metadata *environment.MetaData) {
	// Download localstack config files
	prefix := fmt.Sprintf(localstackS3Key, metadata.CwaCommitSha)
	cxt := context.Background()
	cfg, err := config.LoadDefaultConfig(cxt)
	if err != nil {
		log.Fatalf("Can't get config error: %v", err)
	}
	client := s3.NewFromConfig(cfg)
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(metadata.Bucket),
		Prefix: aws.String(prefix),
	}
	listObjectsOutput, err := client.ListObjectsV2(cxt, listObjectsInput)
	if err != nil {
		log.Fatalf("Got error retrieving list of objects %v", err)
	}
	downloader := manager.NewDownloader(client)
	for _, object := range listObjectsOutput.Contents {
		key := *object.Key
		log.Printf("Download object %s", key)
		keySplit := strings.Split(key, keyDelimiter)
		fileName := keySplit[len(keySplit)-1]
		file, err := os.Create(localstackConfigPath + fileName)
		if err != nil {
			log.Println(err)
		}
		defer file.Close()
		_, err = downloader.Download(cxt, file, &s3.GetObjectInput{
			Bucket: aws.String(metadata.Bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Printf("Error downing file %s error %v", key, err)
		}
	}

	// generate localstack crt files
	writeFile(localstackConfigPath+originalPem, readFile(metadata.CaCertPath))
	writeFile(localstackConfigPath+combinePem, readFile(metadata.CaCertPath))
	writeFile(localstackConfigPath+combinePem, readFile(localstackConfigPath+snakeOilPem))

	// copy crt files to agent directory
	writeFile(tmpDirectory+originalPem, readFile(localstackConfigPath+originalPem))
	writeFile(tmpDirectory+combinePem, readFile(localstackConfigPath+combinePem))
}

func readFile(path string) []byte {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read file %s error %v", path, err)
	}
	return file
}

func writeFile(path string, output []byte) {
	err := os.WriteFile(path, output, 0644)
	if err != nil {
		log.Fatalf("Error writting file %s, error %v,", path, err)
	}
}
