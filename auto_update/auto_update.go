package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
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
		CopyFile(configInput, configOutputPath)
		err := StartAgent(configOutputPath, true)
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

func CopyFile(pathIn string, pathOut string) {
	log.Printf("Copy File %s to %s", pathIn, pathOut)
	pathInAbs, err := filepath.Abs(pathIn)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("File %s abs path %s", pathIn, pathInAbs)
	out, err := exec.Command("bash", "-c", "sudo cp "+pathInAbs+" "+pathOut).Output()

	if err != nil {
		log.Fatal(fmt.Sprint(err) + string(out))
	}

	log.Printf("File : %s copied to : %s", pathIn, pathOut)
}

func StartAgent(configOutputPath string, fatalOnFailure bool) error {
	out, err := exec.
		Command("bash", "-c", "sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a fetch-config -m ec2 -s -c file:"+configOutputPath).
		Output()

	if err != nil && fatalOnFailure {
		log.Fatal(fmt.Sprint(err) + string(out))
	} else if err != nil {
		log.Printf(fmt.Sprint(err) + string(out))
	} else {
		log.Printf("Agent has started")
	}

	return err
}
