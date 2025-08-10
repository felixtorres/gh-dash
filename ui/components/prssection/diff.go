package prssection

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
)

func (m Model) diff() tea.Cmd {
	currRowData := m.GetCurrRow()
	if currRowData == nil {
		return nil
	}

	// Get the current provider and use provider-specific diff command
	provider := data.GetCurrentProvider()
	if provider == nil {
		// Fallback to the original GitHub command
		return m.executeGitHubDiffCommand(currRowData)
	}

	args, err := provider.GetDiffCommand(currRowData.GetNumber(), currRowData.GetRepoNameWithOwner())
	if err != nil {
		return func() tea.Msg {
			return constants.ErrMsg{Err: fmt.Errorf("failed to get diff command: %w", err)}
		}
	}

	if len(args) == 0 {
		return func() tea.Msg {
			return constants.ErrMsg{Err: fmt.Errorf("no diff command available for this provider")}
		}
	}

	// Build command from provider-specific args
	var c *exec.Cmd
	if len(args) == 1 {
		c = exec.Command(args[0])
	} else {
		c = exec.Command(args[0], args[1:]...)
	}
	
	// For GitHub (gh pr diff), set pager environment
	if args[0] == "gh" && len(args) >= 3 && args[1] == "pr" && args[2] == "diff" {
		c.Env = m.Ctx.Config.GetFullScreenDiffPagerEnv()
	}

	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return nil
	})
}

// Fallback GitHub diff command for backward compatibility
func (m Model) executeGitHubDiffCommand(currRowData interface{ GetNumber() int; GetRepoNameWithOwner() string }) tea.Cmd {
	c := exec.Command(
		"gh",
		"pr",
		"diff",
		fmt.Sprint(currRowData.GetNumber()),
		"-R",
		currRowData.GetRepoNameWithOwner(),
	)
	c.Env = m.Ctx.Config.GetFullScreenDiffPagerEnv()

	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return nil
	})
}
