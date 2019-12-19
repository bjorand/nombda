package engine

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	log  = logrus.New()
	runs map[string]*Run
	lock = sync.Mutex{}
)

type Run struct {
	Hook      *Hook
	ID        string
	ExitCode  int
	Completed bool
	Output    string
	Registers map[string]string
}

// type Handler struct {
// 	Name          string `yaml:"name"`
// 	CommandModule `yaml:",inline"`
// 	HandlerName   string `yaml:"handler"`
// }

type Task struct {
	HandlerName          string            `yaml:"handler"`
	Name                 string            `yaml:"name"`
	Command              string            `yaml:"command"`
	Retry                int               `yaml:"retry"`
	Interval             int               `yaml:"interval"`
	Timeout              int               `yaml:"timeout"`
	OnFailure            string            `yaml:"on_failure"`
	ContinueAfterFailure bool              `yaml:"continue_after_failure"`
	OnlyIf               string            `yaml:"only_if"`
	Register             string            `yaml:"register"`
	Vars                 map[string]string `yaml:"vars"`
	Cd                   string            `yaml:"cd"`
}

type Hook struct {
	Name     string
	Action   string
	Handlers map[string][]*Task `yaml:"handlers"`
	Tasks    []*Task            `yaml:"tasks"`
	Runs     []*Run
}

// type HookStep struct {
// 	Name          string `yaml:"name"`
// 	CommandModule `yaml:",inline"`
// 	HandlerName   string `yaml:"handler"`
// }

type HookStepRunResponse struct {
	Stdout []byte
	Stderr []byte
}

type HookEngine struct {
	ConfigDir string
}

func NewHookEngine(configDir string) *HookEngine {
	runs = make(map[string]*Run, 1024)
	return &HookEngine{
		ConfigDir: configDir,
	}
}

func (e *HookEngine) Hooks() ([]*Hook, error) {
	actionsFilename, err := filepath.Glob(e.ConfigDir + "/*/*.yml")
	if err != nil {
		return nil, err
	}

	var hooks []*Hook

	for _, actionFilename := range actionsFilename {
		actionFilenameSplitted := strings.Split(actionFilename, "/")
		if len(actionFilenameSplitted) < 2 {
			return nil, fmt.Errorf("Action found in invalid path %s", actionFilename)
		}
		id := actionFilenameSplitted[len(actionFilenameSplitted)-2]
		actionFileName := filepath.Base(actionFilename)
		extension := filepath.Ext(actionFileName)
		action := actionFileName[0 : len(actionFileName)-len(extension)]
		hook, err := ReadHook(e.ConfigDir, id, action)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}

	return hooks, nil
}

func ReadHookFromFile(p string) (*Hook, error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	h := &Hook{}
	if err := yaml.UnmarshalStrict(data, &h); err != nil {
		return nil, fmt.Errorf("Unable to validate yaml file: %s", err.Error())
	}
	return h, nil
}

func ReadHook(path string, name string, action string) (*Hook, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s.yml", path, name, action))
	if err != nil {
		return nil, err
	}
	h := &Hook{}
	if err := yaml.UnmarshalStrict(data, &h); err != nil {
		return nil, fmt.Errorf("Unable to validate yaml file: %s", err.Error())
	}
	h.Name = name
	h.Action = action
	return h, nil
}

func localRun(command string, envs map[string]string, cd string) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	cmd.Env = os.Environ()
	if cd != "" {
		cmd.Dir = cd
	}
	for k, v := range envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// stderr, err := cmd.StderrPipe()
	// if err != nil {
	// 	return nil, nil, cmd.ProcessState.ExitCode(), err
	// }
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	return nil, nil, cmd.ProcessState.ExitCode(), err
	// }
	// if err := cmd.Start(); err != nil {
	// 	return nil, nil, cmd.ProcessState.ExitCode(), err
	// }
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, cmd.ProcessState.ExitCode(), err
	}

	return output, cmd.ProcessState.ExitCode(), nil
}

// func (h *Taks) Run() ([]byte, int, error) {
// 	outputBytes, exitCode, err := localRun(h.Command, h.Vars, h.Cd)
// 	if err != nil {
// 		return outputBytes, exitCode, err
// 	}
// 	return outputBytes, exitCode, nil
// }

func NewRun(h *Hook) (*Run, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	run := &Run{
		Hook:      h,
		ID:        id.String(),
		Registers: make(map[string]string),
	}
	// if
	// h.Runs = make([]*Run, 0)
	return run, nil
}

func (r *Run) logInfo(input ...string) {
	r.Output += fmt.Sprintf("[INFO] %s\n", strings.Join(input, " "))
}

func (r *Run) logError(input ...string) {
	r.Output += fmt.Sprintf("[ERROR] %s\n", strings.Join(input, " "))
}

func (r *Run) Interpolate(input string, vars map[string]string) string {
	replacers := make([]string, len(r.Registers)*2)
	if len(r.Registers) > 0 {
		for k, v := range r.Registers {
			replacers = append(replacers, fmt.Sprintf("${var.%s}", k))
			replacers = append(replacers, v)
		}
	}

	for k, v := range vars {
		replacers = append(replacers, fmt.Sprintf("${var.%s}", k))
		replacers = append(replacers, v)
	}
	re := strings.NewReplacer(replacers...)
	return re.Replace(input)
}

func (r *Run) RunHandler(src *Task, handlerName string) error {
	r.logInfo("Running handler", handlerName)
	handlerTasks := r.Hook.Handlers[handlerName]
	if handlerTasks == nil {
		r.logError("Unknown handler", handlerName)
		return fmt.Errorf("Unknown handler %s", handlerName)
	}
	for _, handlerTask := range handlerTasks {
		if handlerTask.Vars == nil {
			handlerTask.Vars = make(map[string]string)
		}
		for k, v := range src.Vars {
			handlerTask.Vars[k] = r.Interpolate(v, r.Registers)
		}
		err := r.RunTask(handlerTask)
		if err != nil {
			r.logError("Failure in handler", handlerName)
			if handlerTask.OnFailure != "" {
				r.logInfo("Recovering error in handler", handlerName, "with handler", handlerTask.OnFailure)
				err := r.RunHandler(handlerTask, handlerTask.OnFailure)
				if err != nil {
					if !handlerTask.ContinueAfterFailure {
						return err
					}
					return nil

				}
			}
			return fmt.Errorf("Failure in handler %s", handlerName)
		}
	}
	return nil
}

func (r *Run) RunTask(t *Task) error {
	// only_if is be the first condition
	if t.OnlyIf != "" {
		output, exitCode, err := localRun(r.Interpolate(t.OnlyIf, t.Vars), nil, t.Cd)
		r.ExitCode = exitCode
		r.Output += string(output)
		if err != nil {
			r.logInfo("Skipping step", t.Name)
			return nil
		}
	}
	// run handler module
	if t.HandlerName != "" {
		err := r.RunHandler(t, t.HandlerName)
		if err != nil {
			if t.OnFailure != "" {
				r.logInfo("Recovering error in handler", t.HandlerName, "with handler", t.OnFailure)
				err := r.RunHandler(t, t.OnFailure)
				if err != nil {
					if !t.ContinueAfterFailure {
						return err
					}
					r.logInfo("Continue after failure of handler", t.OnFailure)
					return nil
				}
			} else {
				if !t.ContinueAfterFailure {
					return err
				}
				return nil
			}
		}

	}
	// run command module
	if t.Command != "" {
		r.logInfo("Step command", t.Name)
		output, exitCode, err := localRun(r.Interpolate(t.Command, t.Vars), nil, t.Cd)
		r.ExitCode = exitCode
		r.Output += string(output)
		if t.Register != "" {
			r.Registers[t.Register] = strings.TrimSpace(string(output))
		}
		// command is in error
		// call handler to catch error
		if err != nil {
			// run on_failure handler if any
			if t.OnFailure != "" {
				// 	handlers := r.Hook.Handlers[t.OnFailure]
				// 	if len(handlers) == 0 {
				// 		r.logInfo("Handler [%s] not found", t.OnFailure)
				// 		return fmt.Errorf("Handler [%s] not found", t.OnFailure)
				// 	}
				// 	for _, handler := range handlers {
				// 		err := r.RunTask(handler)
				// 		if err != nil {
				// 			if !t.ContinueAfterFailure {
				// 				return err
				// 			}
				// 			r.logInfo("Continue after failure of handler", t.OnFailure)
				// 			return nil
				// 		}
				// 	}
			}
			return err
		}
	}
	return nil
	// 	// if step is a handler, runs it

	// 	}
	//
	// 	// if step as a command instruction runs it
	// 	if step.Command != "" {
	// 		run.logInfo("Step command", step.Name)
	// 		output, exitCode, err := localRun(run.CommandParser(step.Command), nil, step.Cd)
	// 		run.ExitCode = exitCode
	// 		run.Output += string(output)
	// 		if step.Register != "" {
	// 			run.Registers[step.Register] = strings.TrimSpace(string(output))
	// 		}
	// 		// command is failure
	// 		if err != nil {
	//
	// 			// run on_failure handler if any
	// 			if step.OnFailure != "" {
	//
	// 				if !step.ContinueAfterFailure {
	// 					// return fmt.Errorf("Quit after handler [%s] successful", step.OnFailure)
	// 					run.logError("Failure for step", step.Name)
	// 					return
	// 				}
	// 				run.logInfo("Skip failure after step", step.Name)
	// 				continue
	// 			}
	// 			return
	// 		}
	// 	}
	// }
}

func (h *Hook) AsyncRun(run *Run) {
	defer func() {
		run.Completed = true
		run.logInfo(fmt.Sprintf("Job %s completed with exit code %d", run.ID, run.ExitCode))
	}()
	run.logInfo("Starting job", run.ID)
	for _, task := range h.Tasks {
		err := run.RunTask(task)
		if err != nil {
			if task.ContinueAfterFailure {
				run.logInfo("Continue after failure of task", task.Name)
				continue
			}
			return
		}
	}
}
func (r *Run) Log() string {
	return r.Output
}

func (h *Hook) Run() (*Run, error) {
	run, err := NewRun(h)
	if err != nil {
		return nil, err
	}
	runs[run.ID] = run
	go h.AsyncRun(run)
	return run, nil

}

func (h *Hook) GetRun(id string) (*Run, error) {
	run, ok := runs[id]
	if !ok {
		return nil, fmt.Errorf("run id not found")
	}
	return run, nil
}
