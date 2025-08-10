package prssection

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/common"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func (m *Model) checkout() (tea.Cmd, error) {
	pr := m.GetCurrRow()
	if pr == nil {
		return nil, errors.New("No pr selected")
	}

	repoName := pr.GetRepoNameWithOwner()
	repoPath, ok := common.GetRepoLocalPath(repoName, m.Ctx.Config.RepoPaths)

	if !ok {
		return nil, errors.New("Local path to repo not specified, set one in your config.yml under repoPaths")
	}

	prNumber := pr.GetNumber()
	taskId := fmt.Sprintf("checkout_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Checking out PR #%d", prNumber),
		FinishedText: fmt.Sprintf("PR #%d has been checked out at %s", prNumber, repoPath),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		// Get the current provider and use provider-specific checkout command
		provider := data.GetCurrentProvider()
		var c *exec.Cmd
		var err error
		
		if provider != nil {
			// Use provider-specific command
			args, cmdErr := provider.GetCheckoutCommand(prNumber, repoName)
			if cmdErr != nil {
				return constants.TaskFinishedMsg{TaskId: taskId, Err: cmdErr}
			}
			
			if len(args) == 0 {
				return constants.TaskFinishedMsg{TaskId: taskId, Err: fmt.Errorf("checkout command not available for this provider")}
			}
			
			// Build command from provider-specific args
			if len(args) == 1 {
				c = exec.Command(args[0])
			} else {
				c = exec.Command(args[0], args[1:]...)
			}
		} else {
			// Fallback to original GitHub command
			c = exec.Command(
				"gh",
				"pr",
				"checkout",
				fmt.Sprint(prNumber),
			)
		}
		
		userHomeDir, _ := os.UserHomeDir()
		if strings.HasPrefix(repoPath, "~") {
			repoPath = strings.Replace(repoPath, "~", userHomeDir, 1)
		}

		c.Dir = repoPath
		err = c.Run()
		return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
	}), nil
}
