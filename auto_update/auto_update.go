package main

import (
	"github.com/aws/amazon-cloudwatch-agent-test/internal/common"
	"log"
	"os/exec"
	"time"
)

const (
	configOutputPath = "/opt/aws/amazon-cloudwatch-agent/bin/config.json"
	configInput      = "resources/elephantConfig.json"
)

func main() {
	for true {
		syncAgent()
		installAgent()
		common.CopyFile(configInput, configOutputPath)
		err := common.StartAgent(configOutputPath, true)
		if err != nil {
			log.Printf("Err starting agent %v", err)
		}
		time.Sleep(time.Hour)
		log.Printf("End auto update process going to sleep for 1 hr")
	}
}

func syncAgent() {
	err := exec.Command("bash", "-c", "aws s3 sync s3://private-cloudwatch-agent-integration-test/release/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm .")
	if err != nil {
		log.Printf("Failed to download agent err %v", err)
	}
}

func installAgent() {
	err := exec.Command("bash", "-c", "sudo yum install -y amazon-cloudwatch-agent.rpm")
	if err != nil {
		log.Printf("Failed to install agent err %v", err)
	}
}
