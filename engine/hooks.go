package engine

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	log = logrus.New()
)

type Handler struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Vars    map[string]string `yaml:"vars"`
}

type Command struct {
}

type Hook struct {
	Handlers map[string][]*Handler `yaml:"handlers"`
	Steps    []*HookStep           `yaml:"steps"`
}

type HookStep struct {
	Name                 string `yaml:"name"`
	Command              string `yaml:"command"`
	Retry                int    `yaml:"retry"`
	Interval             int    `yaml:"interval"`
	Timeout              int    `yaml:"timeout"`
	OnFailure            string `yaml:"on_failure"`
	ContinueAfterFailure bool   `yaml:"continue_after_failure"`
	OnlyIf               string `yaml:"only_if"`
	Register             string `yaml:"register"`
	Response             *HookStepRunResponse
}

type HookStepRunResponse struct {
	Stdout []byte
	Stderr []byte
}

func ReadHook(id string, action string) (*Hook, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("tmp/%s/%s.yml", id, action))
	if err != nil {
		return nil, err
	}
	h := &Hook{}
	if err := yaml.UnmarshalStrict(data, &h); err != nil {
		return nil, fmt.Errorf("Unable to validate yaml file: %s", err.Error())
	}
	return h, nil
}

func localRun(command string, envs map[string]string) ([]byte, error) {
	fmt.Println(command, envs)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	cmd.Env = os.Environ()
	for k, v := range envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	stderrBytes, _ := ioutil.ReadAll(stderr)
	stdoutBytes, _ := ioutil.ReadAll(stdout)
	fmt.Println("STDOUT", string(stdoutBytes))
	fmt.Println("STDERR", string(stderrBytes))
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%s : %s %s (%s)", command, string(stdoutBytes), string(stderrBytes), err.Error())
		// step.Response.Stderr = []byte(errRun.Error())
		// continue
	}
	return stdoutBytes, nil
}

func (h *Handler) Run() ([]byte, error) {
	stdoutBytes, err := localRun(h.Command, h.Vars)
	if err != nil {
		return nil, err
	}
	return stdoutBytes, nil
}
func (h *Hook) Run() error {
	for _, step := range h.Steps {
		step.Response = &HookStepRunResponse{}

		if step.OnlyIf != "" {
			_, err := localRun(step.OnlyIf, nil)
			if err != nil {
				log.Infof("Skipping %s", step.Name)
				continue
			}
		}
		log.Infof("Running %s", step.Name)
		_, err := localRun(step.Command, nil)
		if err != nil {
			if step.OnFailure != "" {
				handlers := h.Handlers[step.OnFailure]
				for _, handler := range handlers {
					_, err := handler.Run()
					if err != nil {
						return fmt.Errorf("Handler [%s] %s failed: %s", step.OnFailure, handler.Name, err.Error())
					}
				}
				if !step.ContinueAfterFailure {
					return fmt.Errorf("Handler [%s] successful", step.OnFailure)
				}
				continue
			}
			return err
		}
	}
	return nil
}
