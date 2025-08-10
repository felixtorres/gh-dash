package prssection

import (
	"bytes"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/gen2brain/beeep"

	"github.com/dlvhdr/gh-dash/v4/data"
	prComponent "github.com/dlvhdr/gh-dash/v4/ui/components/pr"
	"github.com/dlvhdr/gh-dash/v4/ui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func (m *Model) watchChecks() tea.Cmd {
	pr := m.GetCurrRow()
	if pr == nil {
		return nil
	}

	prNumber := pr.GetNumber()
	title := pr.GetTitle()
	url := pr.GetUrl()
	repoNameWithOwner := pr.GetRepoNameWithOwner()
	taskId := fmt.Sprintf("pr_reopen_%d", prNumber)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Watching checks for PR #%d", prNumber),
		FinishedText: fmt.Sprintf("Watching checks for PR #%d", prNumber),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		// Get the current provider and use provider-specific watch checks command
		provider := data.GetCurrentProvider()
		var c *exec.Cmd
		
		if provider != nil {
			// Use provider-specific command
			args, cmdErr := provider.GetWatchChecksCommand(prNumber, repoNameWithOwner)
			if cmdErr != nil {
				return constants.TaskFinishedMsg{
					SectionId:   m.Id,
					SectionType: SectionType,
					TaskId:      taskId,
					Err:         cmdErr,
					Msg: tasks.UpdatePRMsg{PrNumber: prNumber},
				}
			}
			
			if len(args) == 0 {
				return constants.TaskFinishedMsg{
					SectionId:   m.Id,
					SectionType: SectionType,
					TaskId:      taskId,
					Err:         fmt.Errorf("watch checks command not available for this provider"),
					Msg:         tasks.UpdatePRMsg{PrNumber: prNumber},
				}
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
				"checks",
				"--watch",
				"--fail-fast",
				fmt.Sprint(prNumber),
				"-R",
				repoNameWithOwner,
			)
		}

		var outb, errb bytes.Buffer
		c.Stdout = &outb
		c.Stderr = &errb

		err := c.Start()
		go func() {
			err := c.Wait()
			if err != nil {
				log.Debug("Error waiting for watch command to finish", "err", err, "stderr", errb.String(), "stdout", outb.String())
			}

			// TODO: check for installation of terminal-notifier or alternative as logo isn't supported
			updatedPr, err := data.FetchPullRequest(url)
			if err != nil {
				log.Debug("Error fetching updated PR details", "url", url, "err", err)
			}

			renderedPr := prComponent.PullRequest{Ctx: m.Ctx, Data: &updatedPr}
			checksRollup := " Checks are pending"
			switch renderedPr.GetStatusChecksRollup() {
			case "SUCCESS":
				checksRollup = "✅ Checks have passed"
			case "FAILURE":
				checksRollup = "❌ Checks have failed"
			}

			err = beeep.Notify(
				fmt.Sprintf("gh-dash: %s", title),
				fmt.Sprintf("PR #%d in %s\n%s", prNumber, repoNameWithOwner, checksRollup),
				"",
			)
			if err != nil {
				log.Debug("Error showing system notification", "err", err)
			}
		}()

		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: tasks.UpdatePRMsg{
				PrNumber: prNumber,
			},
		}
	})
}
