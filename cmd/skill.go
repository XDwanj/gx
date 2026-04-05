package cmd

import (
	"github.com/XDwanj/gx/internal/skill"

	"github.com/spf13/cobra"
)

func newSkillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "skill",
		Short: "Print the agent skill file to stdout",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := cmd.OutOrStdout().Write([]byte(skill.Text))
			return err
		},
	}
}
