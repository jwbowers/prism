package cli

import (
	"fmt"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/spf13/cobra"
)

// ProjectCobraCommands handles project management commands
type ProjectCobraCommands struct {
	app *App
}

// NewProjectCobraCommands creates new project cobra commands
func NewProjectCobraCommands(app *App) *ProjectCobraCommands {
	return &ProjectCobraCommands{app: app}
}

// CreateProjectCommand creates the project command with subcommands
func (pc *ProjectCobraCommands) CreateProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Prism projects",
		Long: `Manage projects for organizing instances, tracking budgets, and collaborating
with team members.`,
	}

	// Add subcommands
	cmd.AddCommand(
		pc.createListCommand(),
		pc.createCreateCommand(),
		pc.createInfoCommand(),
		pc.createDeleteCommand(),
		pc.createMembersCommand(),
		pc.createBudgetCommand(),
		pc.createInstancesCommand(),
		pc.createTemplatesCommand(),
		pc.createInviteCommand(),
		pc.createInvitationsCommand(),
		// v0.12.0 governance subcommands
		pc.createQuotaCommand(),
		pc.createGrantPeriodCommand(),
		pc.createOnboardingCommand(),
	)

	return cmd
}

// createListCommand creates the list subcommand
func (pc *ProjectCobraCommands) createListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pc.app.Project([]string{"list"})
		},
	}
}

// createCreateCommand creates the create subcommand
func (pc *ProjectCobraCommands) createCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			description, _ := cmd.Flags().GetString("description")
			budget, _ := cmd.Flags().GetFloat64("budget")
			owner, _ := cmd.Flags().GetString("owner")

			createArgs := []string{"create", args[0]}
			if description != "" {
				createArgs = append(createArgs, "--description", description)
			}
			if budget > 0 {
				createArgs = append(createArgs, "--budget", fmt.Sprintf("%.2f", budget))
			}
			if owner != "" {
				createArgs = append(createArgs, "--owner", owner)
			}

			return pc.app.Project(createArgs)
		},
	}

	cmd.Flags().String("description", "", "Project description")
	cmd.Flags().Float64("budget", 0, "Budget limit")
	cmd.Flags().String("owner", "", "Project owner")

	return cmd
}

// createInfoCommand creates the info subcommand
func (pc *ProjectCobraCommands) createInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show detailed project information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pc.app.Project([]string{"info", args[0]})
		},
	}
}

// createDeleteCommand creates the delete subcommand
func (pc *ProjectCobraCommands) createDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pc.app.Project([]string{"delete", args[0]})
		},
	}
}

// createMembersCommand creates the members management subcommand
func (pc *ProjectCobraCommands) createMembersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members <project>",
		Short: "Manage project members",
		Args:  cobra.ExactArgs(1),
	}

	// Members subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "add <email> <role>",
			Short: "Add a member to project",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				projectName := cmd.Parent().Flag("project").Value.String()
				if projectName == "" {
					// Get project from parent args
					parentArgs := cmd.Parent().Flags().Args()
					if len(parentArgs) > 0 {
						projectName = parentArgs[0]
					}
				}
				return pc.app.Project([]string{"members", projectName, "add", args[0], args[1]})
			},
		},
		&cobra.Command{
			Use:   "remove <email>",
			Short: "Remove a member from project",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				projectName := cmd.Parent().Flag("project").Value.String()
				if projectName == "" {
					// Get project from parent args
					parentArgs := cmd.Parent().Flags().Args()
					if len(parentArgs) > 0 {
						projectName = parentArgs[0]
					}
				}
				return pc.app.Project([]string{"members", projectName, "remove", args[0]})
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List project members",
			RunE: func(cmd *cobra.Command, args []string) error {
				projectName := cmd.Parent().Flag("project").Value.String()
				if projectName == "" {
					// Get project from parent args
					parentArgs := cmd.Parent().Flags().Args()
					if len(parentArgs) > 0 {
						projectName = parentArgs[0]
					}
				}
				return pc.app.Project([]string{"members", projectName})
			},
		},
	)

	return cmd
}

// createBudgetCommand creates the budget management subcommand
func (pc *ProjectCobraCommands) createBudgetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Manage project budgets and cost tracking",
		Long: `Configure project budgets, set spending limits, configure alerts,
and enable cost tracking for research projects.`,
	}

	// Budget subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "status <project>",
			Short: "Show budget status and spending",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return pc.app.Project([]string{"budget", "status", args[0]})
			},
		},
		&cobra.Command{
			Use:   "set <project> <amount>",
			Short: "Set or enable project budget",
			Long:  `Set a budget for a project and enable cost tracking. Amount should be in USD.`,
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return pc.app.Project([]string{"budget", "set", args[0], args[1]})
			},
		},
		&cobra.Command{
			Use:   "disable <project>",
			Short: "Disable cost tracking for project",
			Long:  `Disable budget tracking and cost monitoring for a project.`,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return pc.app.Project([]string{"budget", "disable", args[0]})
			},
		},
		&cobra.Command{
			Use:   "history <project>",
			Short: "Show budget spending history",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return pc.app.Project([]string{"budget", "history", args[0]})
			},
		},
		&cobra.Command{
			Use:   "prevent-launches <project>",
			Short: "Block new instance launches for project",
			Long: `Prevent new instance launches for a project due to budget limits.
This is typically used when a project reaches its budget hard cap.`,
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return pc.app.Project([]string{"budget", "prevent-launches", args[0]})
			},
		},
		&cobra.Command{
			Use:   "allow-launches <project>",
			Short: "Allow instance launches for project",
			Long: `Remove launch prevention for a project, allowing new instances to be created.
This clears the budget hard cap block temporarily.`,
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return pc.app.Project([]string{"budget", "allow-launches", args[0]})
			},
		},
	)

	return cmd
}

// createInstancesCommand creates the instances subcommand
func (pc *ProjectCobraCommands) createInstancesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "instances <project>",
		Short: "List instances in a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pc.app.Project([]string{"instances", args[0]})
		},
	}
}

// createTemplatesCommand creates the templates subcommand
func (pc *ProjectCobraCommands) createTemplatesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "templates <project>",
		Short: "List templates in a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pc.app.Project([]string{"templates", args[0]})
		},
	}
}

// ============================================================================
// v0.12.0: Quota, Grant Period, Onboarding commands
// ============================================================================

// createQuotaCommand creates the `prism project quota` subcommand (#151)
func (pc *ProjectCobraCommands) createQuotaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quota <project>",
		Short: "Manage resource quotas by role",
		Long: `View and set per-role resource quotas for a project.

Quotas restrict how many instances members of a given role can launch
and what instance types they are allowed to use.

Examples:
  prism project quota get my-project
  prism project quota set my-project --role student --max-instances 2
  prism project quota set my-project --role member --max-instance-type t3 --max-spend-daily 5`,
		Args: cobra.ExactArgs(1),
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "get",
			Short: "Show current role quotas",
			RunE: func(cmd *cobra.Command, args []string) error {
				projectName := cmd.Parent().Flags().Arg(0)
				if projectName == "" {
					return fmt.Errorf("project name required")
				}

				proj, err := pc.app.apiClient.GetProject(pc.app.ctx, projectName)
				if err != nil {
					return err
				}

				quotas, err := pc.app.apiClient.GetProjectQuotas(pc.app.ctx, proj.ID)
				if err != nil {
					return err
				}

				if len(quotas) == 0 {
					fmt.Printf("No quotas configured for project %s\n", projectName)
					return nil
				}

				fmt.Printf("Role quotas for %s:\n", projectName)
				for _, q := range quotas {
					fmt.Printf("  %-12s  max-instances=%d  max-type=%q  max-daily=$%.2f\n",
						q.Role, q.MaxInstances, q.MaxInstanceType, q.MaxSpendDaily)
				}
				return nil
			},
		},
	)

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set a role quota",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := cmd.Parent().Flags().Arg(0)
			if projectName == "" {
				return fmt.Errorf("project name required")
			}

			role, _ := cmd.Flags().GetString("role")
			maxInstances, _ := cmd.Flags().GetInt("max-instances")
			maxType, _ := cmd.Flags().GetString("max-instance-type")
			maxDaily, _ := cmd.Flags().GetFloat64("max-spend-daily")

			if role == "" {
				return fmt.Errorf("--role is required")
			}

			proj, err := pc.app.apiClient.GetProject(pc.app.ctx, projectName)
			if err != nil {
				return err
			}

			quota := types.RoleQuota{
				Role:            types.ProjectRole(role),
				MaxInstances:    maxInstances,
				MaxInstanceType: maxType,
				MaxSpendDaily:   maxDaily,
			}

			if err := pc.app.apiClient.SetProjectQuota(pc.app.ctx, proj.ID, quota); err != nil {
				return err
			}

			fmt.Printf("✅ Quota set for role %q in project %s\n", role, projectName)
			return nil
		},
	}
	setCmd.Flags().String("role", "", "Project role (owner, admin, member, viewer)")
	setCmd.Flags().Int("max-instances", -1, "Maximum concurrent instances (-1 = unlimited)")
	setCmd.Flags().String("max-instance-type", "", "Instance type prefix limit (e.g. t3)")
	setCmd.Flags().Float64("max-spend-daily", 0, "Daily spend limit in USD (0 = unlimited)")
	cmd.AddCommand(setCmd)

	return cmd
}

// createGrantPeriodCommand creates the `prism project grant-period` subcommand (#152)
func (pc *ProjectCobraCommands) createGrantPeriodCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant-period <project>",
		Short: "Manage grant period and auto-freeze",
		Long: `Configure a funding grant period for a project.
When AutoFreeze is enabled, the project is automatically paused when the grant period ends.

Examples:
  prism project grant-period set my-project --start 2026-01-01 --end 2026-12-31 --auto-freeze
  prism project grant-period status my-project
  prism project grant-period delete my-project`,
		Args: cobra.ExactArgs(1),
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "status",
			Short: "Show grant period status",
			RunE: func(cmd *cobra.Command, args []string) error {
				projectName := cmd.Parent().Flags().Arg(0)
				if projectName == "" {
					return fmt.Errorf("project name required")
				}

				proj, err := pc.app.apiClient.GetProject(pc.app.ctx, projectName)
				if err != nil {
					return err
				}

				gp, err := pc.app.apiClient.GetGrantPeriod(pc.app.ctx, proj.ID)
				if err != nil {
					return err
				}

				if gp == nil {
					fmt.Printf("No grant period configured for %s\n", projectName)
					return nil
				}

				fmt.Printf("Grant Period: %s\n", gp.Name)
				fmt.Printf("  Start:       %s\n", gp.StartDate.Format("2006-01-02"))
				fmt.Printf("  End:         %s\n", gp.EndDate.Format("2006-01-02"))
				fmt.Printf("  Auto-Freeze: %v\n", gp.AutoFreeze)
				if gp.FrozenAt != nil {
					fmt.Printf("  Frozen At:   %s\n", gp.FrozenAt.Format("2006-01-02 15:04"))
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "delete",
			Short: "Remove the grant period",
			RunE: func(cmd *cobra.Command, args []string) error {
				projectName := cmd.Parent().Flags().Arg(0)
				if projectName == "" {
					return fmt.Errorf("project name required")
				}

				proj, err := pc.app.apiClient.GetProject(pc.app.ctx, projectName)
				if err != nil {
					return err
				}

				if err := pc.app.apiClient.DeleteGrantPeriod(pc.app.ctx, proj.ID); err != nil {
					return err
				}

				fmt.Printf("✅ Grant period removed from %s\n", projectName)
				return nil
			},
		},
	)

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set or update the grant period",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := cmd.Parent().Flags().Arg(0)
			if projectName == "" {
				return fmt.Errorf("project name required")
			}

			name, _ := cmd.Flags().GetString("name")
			startStr, _ := cmd.Flags().GetString("start")
			endStr, _ := cmd.Flags().GetString("end")
			autoFreeze, _ := cmd.Flags().GetBool("auto-freeze")

			if startStr == "" || endStr == "" {
				return fmt.Errorf("--start and --end are required")
			}

			start, err := time.Parse("2006-01-02", startStr)
			if err != nil {
				return fmt.Errorf("invalid --start date (use YYYY-MM-DD): %w", err)
			}

			end, err := time.Parse("2006-01-02", endStr)
			if err != nil {
				return fmt.Errorf("invalid --end date (use YYYY-MM-DD): %w", err)
			}

			proj, err := pc.app.apiClient.GetProject(pc.app.ctx, projectName)
			if err != nil {
				return err
			}

			gp := types.GrantPeriod{
				Name:       name,
				StartDate:  start,
				EndDate:    end,
				AutoFreeze: autoFreeze,
			}

			if err := pc.app.apiClient.SetGrantPeriod(pc.app.ctx, proj.ID, gp); err != nil {
				return err
			}

			fmt.Printf("✅ Grant period set for %s\n", projectName)
			fmt.Printf("   %s → %s (auto-freeze: %v)\n",
				start.Format("2006-01-02"), end.Format("2006-01-02"), autoFreeze)
			return nil
		},
	}
	setCmd.Flags().String("name", "Grant Period", "Name for the grant period")
	setCmd.Flags().String("start", "", "Start date (YYYY-MM-DD)")
	setCmd.Flags().String("end", "", "End date (YYYY-MM-DD)")
	setCmd.Flags().Bool("auto-freeze", false, "Automatically pause project when grant period ends")
	cmd.AddCommand(setCmd)

	return cmd
}

// createOnboardingCommand creates the `prism project onboarding` subcommand (#154)
func (pc *ProjectCobraCommands) createOnboardingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "onboarding <project>",
		Short: "Manage new member onboarding templates",
		Long: `Configure onboarding templates that are automatically applied when new members join.

Examples:
  prism project onboarding list my-project
  prism project onboarding add my-project --name grad-student --budget 200`,
		Args: cobra.ExactArgs(1),
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List onboarding templates",
			RunE: func(cmd *cobra.Command, args []string) error {
				projectName := cmd.Parent().Flags().Arg(0)
				if projectName == "" {
					return fmt.Errorf("project name required")
				}

				proj, err := pc.app.apiClient.GetProject(pc.app.ctx, projectName)
				if err != nil {
					return err
				}

				if len(proj.OnboardingTemplates) == 0 {
					fmt.Printf("No onboarding templates configured for %s\n", projectName)
					return nil
				}

				fmt.Printf("Onboarding templates for %s:\n", projectName)
				for _, t := range proj.OnboardingTemplates {
					fmt.Printf("  %-20s  budget=$%.2f  templates=%v\n", t.Name, t.BudgetLimit, t.Templates)
				}
				return nil
			},
		},
	)

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add or update an onboarding template",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := cmd.Parent().Flags().Arg(0)
			if projectName == "" {
				return fmt.Errorf("project name required")
			}

			name, _ := cmd.Flags().GetString("name")
			budget, _ := cmd.Flags().GetFloat64("budget")
			templatesStr, _ := cmd.Flags().GetStringSlice("templates")

			if name == "" {
				return fmt.Errorf("--name is required")
			}

			proj, err := pc.app.apiClient.GetProject(pc.app.ctx, projectName)
			if err != nil {
				return err
			}

			tmpl := types.OnboardingTemplate{
				Name:        name,
				BudgetLimit: budget,
				Templates:   templatesStr,
			}

			if err := pc.app.apiClient.AddOnboardingTemplate(pc.app.ctx, proj.ID, tmpl); err != nil {
				return err
			}

			fmt.Printf("✅ Onboarding template %q added to %s\n", name, projectName)
			return nil
		},
	}
	addCmd.Flags().String("name", "", "Onboarding template name")
	addCmd.Flags().Float64("budget", 0, "Per-member budget allocation on join")
	addCmd.Flags().StringSlice("templates", nil, "Workspace templates to provision")
	cmd.AddCommand(addCmd)

	return cmd
}
