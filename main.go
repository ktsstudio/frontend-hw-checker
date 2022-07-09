package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"
	"validator/v1/pkg/app_config"
	// "validator/v1/pkg/s3_uploader"
	"gopkg.in/yaml.v2"
)

var PytestResultPattern = regexp.MustCompile("={25}\\s*(?P<failed>\\d+ failed,?)?\\s*(?P<passed>\\d+ passed,?)?\\s*(?P<skipped>\\d+ skipped,?)? in .*={25}")

func getStudentConfig() error {
	f, err := os.Open(config.StudentConfig.ConfigFilename)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		log.Printf("Can not open %s", config.StudentConfig.ConfigFilename)
		return err
	}

	log.Printf("Reading %s", config.StudentConfig.ConfigFilename)
	buffer, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Can not read %s", config.StudentConfig.ConfigFilename)
		return err
	}

	studentConfig := app_config.StudentConfig{}

	err = yaml.Unmarshal(buffer, &studentConfig)
	log.Printf("Parse %s", config.StudentConfig.ConfigFilename)
	if err != nil {
		log.Printf("Can not parse %s", config.StudentConfig.ConfigFilename)
		return err
	}
	if studentConfig.UserToken == "" {
		log.Printf("Empty user token")
		return errors.New(fmt.Sprintf("can not find user_token in %s", config.StudentConfig.ConfigFilename))
	}
	config.StudentConfig.UserToken = studentConfig.UserToken
	return nil
}

func submitResult() error {
	client := http.Client{}

	type RequestExtra struct {
		StudentRepo string `json:"student_repo"`
		StudentRef  string `json:"student_ref"`
	}

	type Request struct {
		Created   time.Time    `json:"created"`
		TaskID    string       `json:"task_id"`
		UserToken string       `json:"user_token"`
		Extra     RequestExtra `json:"extra"`
	}
	data, err := json.Marshal(Request{
		Created:   time.Now(),
		TaskID:    config.CallbackTaskId,
		UserToken: config.StudentConfig.UserToken,
		Extra: RequestExtra{
			StudentRepo: config.StudentConfig.StudentRepo,
			StudentRef:  config.StudentConfig.StudentRef,
		},
	})
	req, err := http.NewRequest("POST", config.LmsBaseUrl, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.LmsCompanyToken))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Response status code %d instead of 200", resp.StatusCode))
	}
	return nil
}

func runCommand(cmd string, args ...string) (bytes.Buffer, error) {
	var outBuffer bytes.Buffer
	execCmd := exec.Command(cmd, args...)
	execCmd.Stdout = &outBuffer
	execCmd.Stderr = &outBuffer

	err := execCmd.Run()
	if outBuffer.Len() > 0 {
		log.Println(outBuffer.String())
	}
	if err != nil {
		return bytes.Buffer{}, err
	}
	return outBuffer, nil
}

func setupEnvironment() error {
	var err error
	log.Printf("Setuping tests environment")
	if _, err = os.Stat("package.json"); err == nil {
		log.Printf("Found package.json")
		log.Printf("Run yarn install")
		_, err = runCommand("yarn", "install")
		if err != nil {
			log.Printf("Can not install dependecies: %v", err)
			return err
		}
	}

	log.Printf("Removing student tests/ dir")
	_, err = runCommand("rm", "-rf", "src/__test__")
	if err != nil {
		log.Printf("Can not remove src/__test__/ folder: %v", err)
		return err
	}

	log.Printf("Cloning original repo")
	_, err = runCommand("git", "clone", config.GithubRepo, "original_repo")
	if err != nil {
		log.Printf("Can not clone original repo: %v", err)
		return err
	}

	log.Printf("Moving tests folder to student code")
	_, err = runCommand("mv", "original_repo/src/__test__", "src/__test__/")
	if err != nil {
		log.Printf("Can not move original src/__test__/ folder to student code")
		return err
	}
	return nil
}

func runTests() error {
	_, err := runCommand("yarn", "test")
	if err != nil {
		log.Printf("Tests failed: %v", err)
		return err
	}
	return nil
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Failed to run")
		}
	}()

	config.StudentConfig.StudentRepo = os.Getenv("GITHUB_REPOSITORY")
	config.StudentConfig.StudentRef = os.Getenv("GITHUB_REF")
	if config.StudentConfig.StudentRepo == "" || config.StudentConfig.StudentRef == "" {
		log.Fatal("No info about GitHub Repo is supplied")
	}

	// uploader := s3_uploader.NewS3Uploader(config)
	// isCorrect := false
	// defer func() {
	// 	log.Print("Uploading source code")
	// 	err := uploader.UploadRepo(isCorrect)
	// 	if err != nil {
	// 		log.Fatalf("Can not upload repo to s3: %v", err)
	// 	}
	// }()

	log.Printf("Searching for %s", config.StudentConfig.ConfigFilename)
	err := getStudentConfig()
	if err != nil {
		log.Panicf("Can not read %s: %v", config.StudentConfig.ConfigFilename, err)
	}

	err = setupEnvironment()
	if err != nil {
		log.Panic("Can not setup environment")
	}

	err = runTests()
	if err != nil {
		log.Panic("Can not run tests")
	}

	log.Printf("Submiting success result")
	err = submitResult()
	if err != nil {
		log.Panicf("Can not submit result: %v. Please try again later.", err)
	}
	// isCorrect = true
}
