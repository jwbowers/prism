package cli

import (
	"github.com/spf13/cobra"
)

// CourseCobraCommands handles course management commands.
type CourseCobraCommands struct {
	app *App
}

// NewCourseCobraCommands creates new course cobra commands.
func NewCourseCobraCommands(app *App) *CourseCobraCommands {
	return &CourseCobraCommands{app: app}
}

// CreateCourseCommand creates the 'course' top-level command.
func (cc *CourseCobraCommands) CreateCourseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "course",
		Short: "Manage university courses",
		Long: `Manage courses, rosters, template whitelists, and per-student budgets.

Examples:
  prism course list
  prism course create CS101 --title "Intro to CS" --semester "Fall 2099" \
        --start 2099-09-01 --end 2099-12-15 --owner prof1
  prism course show <id>
  prism course close <id>
  prism course templates add <id> python-ml
  prism course students enroll <id> --email student@uni.edu
  prism course budget distribute <id> --amount 50`,
	}

	cmd.AddCommand(
		cc.createListCommand(),
		cc.createCreateCommand(),
		cc.createShowCommand(),
		cc.createCloseCommand(),
		cc.createDeleteCommand(),
		cc.createTemplatesCommand(),
		cc.createStudentsCommand(),
		cc.createBudgetCommand(),
	)
	return cmd
}

func (cc *CourseCobraCommands) createListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List courses",
		RunE: func(cmd *cobra.Command, args []string) error {
			semester, _ := cmd.Flags().GetString("semester")
			owner, _ := cmd.Flags().GetString("owner")
			status, _ := cmd.Flags().GetString("status")

			cliArgs := []string{"list"}
			if semester != "" {
				cliArgs = append(cliArgs, "--semester", semester)
			}
			if owner != "" {
				cliArgs = append(cliArgs, "--owner", owner)
			}
			if status != "" {
				cliArgs = append(cliArgs, "--status", status)
			}
			return cc.app.Course(cliArgs)
		},
	}
	cmd.Flags().String("semester", "", "Filter by semester (e.g. 'Fall 2099')")
	cmd.Flags().String("owner", "", "Filter by owner user ID")
	cmd.Flags().String("status", "", "Filter by status (active, pending, closed, archived)")
	return cmd
}

func (cc *CourseCobraCommands) createCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <code>",
		Short: "Create a new course",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title, _ := cmd.Flags().GetString("title")
			semester, _ := cmd.Flags().GetString("semester")
			owner, _ := cmd.Flags().GetString("owner")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			budget, _ := cmd.Flags().GetFloat64("budget")
			dept, _ := cmd.Flags().GetString("department")

			cliArgs := []string{"create", args[0]}
			if title != "" {
				cliArgs = append(cliArgs, "--title", title)
			}
			if semester != "" {
				cliArgs = append(cliArgs, "--semester", semester)
			}
			if owner != "" {
				cliArgs = append(cliArgs, "--owner", owner)
			}
			if start != "" {
				cliArgs = append(cliArgs, "--start", start)
			}
			if end != "" {
				cliArgs = append(cliArgs, "--end", end)
			}
			if budget > 0 {
				cliArgs = append(cliArgs, "--budget", cmd.Flags().Lookup("budget").Value.String())
			}
			if dept != "" {
				cliArgs = append(cliArgs, "--department", dept)
			}
			return cc.app.Course(cliArgs)
		},
	}
	cmd.Flags().String("title", "", "Course title")
	cmd.Flags().String("semester", "", "Semester label (e.g. 'Fall 2099')")
	cmd.Flags().String("owner", "", "Instructor user ID")
	cmd.Flags().String("start", "", "Semester start date (YYYY-MM-DD)")
	cmd.Flags().String("end", "", "Semester end date (YYYY-MM-DD)")
	cmd.Flags().Float64("budget", 0, "Per-student budget (USD)")
	cmd.Flags().String("department", "", "Department name")
	return cmd
}

func (cc *CourseCobraCommands) createShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show course details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"show", args[0]})
		},
	}
}

func (cc *CourseCobraCommands) createCloseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "close <id>",
		Short: "Close a course (semester end)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"close", args[0]})
		},
	}
}

func (cc *CourseCobraCommands) createDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a course",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"delete", args[0]})
		},
	}
}

func (cc *CourseCobraCommands) createTemplatesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Manage approved template whitelist for a course",
	}

	listCmd := &cobra.Command{
		Use:   "list <course-id>",
		Short: "List approved templates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"templates", "list", args[0]})
		},
	}

	addCmd := &cobra.Command{
		Use:   "add <course-id> <template-slug>",
		Short: "Add a template to the whitelist",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"templates", "add", args[0], args[1]})
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove <course-id> <template-slug>",
		Short: "Remove a template from the whitelist",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"templates", "remove", args[0], args[1]})
		},
	}

	cmd.AddCommand(listCmd, addCmd, removeCmd)
	return cmd
}

func (cc *CourseCobraCommands) createStudentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "students",
		Short: "Manage course roster",
	}

	listCmd := &cobra.Command{
		Use:   "list <course-id>",
		Short: "List course members",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			role, _ := cmd.Flags().GetString("role")
			cliArgs := []string{"students", "list", args[0]}
			if role != "" {
				cliArgs = append(cliArgs, "--role", role)
			}
			return cc.app.Course(cliArgs)
		},
	}
	listCmd.Flags().String("role", "", "Filter by role (student, ta, instructor)")

	enrollCmd := &cobra.Command{
		Use:   "enroll <course-id>",
		Short: "Enroll a member",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			userID, _ := cmd.Flags().GetString("user")
			name, _ := cmd.Flags().GetString("name")
			role, _ := cmd.Flags().GetString("role")
			budget, _ := cmd.Flags().GetFloat64("budget")

			cliArgs := []string{"students", "enroll", args[0]}
			if email != "" {
				cliArgs = append(cliArgs, "--email", email)
			}
			if userID != "" {
				cliArgs = append(cliArgs, "--user", userID)
			}
			if name != "" {
				cliArgs = append(cliArgs, "--name", name)
			}
			if role != "" {
				cliArgs = append(cliArgs, "--role", role)
			}
			if budget > 0 {
				cliArgs = append(cliArgs, "--budget", cmd.Flags().Lookup("budget").Value.String())
			}
			return cc.app.Course(cliArgs)
		},
	}
	enrollCmd.Flags().String("email", "", "Student email")
	enrollCmd.Flags().String("user", "", "User ID")
	enrollCmd.Flags().String("name", "", "Display name")
	enrollCmd.Flags().String("role", "student", "Role (student, ta, instructor)")
	enrollCmd.Flags().Float64("budget", 0, "Per-student budget override (USD)")

	unenrollCmd := &cobra.Command{
		Use:   "unenroll <course-id> <user-id>",
		Short: "Remove a member from the course",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"students", "unenroll", args[0], args[1]})
		},
	}

	cmd.AddCommand(listCmd, enrollCmd, unenrollCmd)
	return cmd
}

func (cc *CourseCobraCommands) createBudgetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Manage course budget",
	}

	showCmd := &cobra.Command{
		Use:   "show <course-id>",
		Short: "Show budget summary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"budget", "show", args[0]})
		},
	}

	distributeCmd := &cobra.Command{
		Use:   "distribute <course-id>",
		Short: "Set per-student budget",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, _ := cmd.Flags().GetString("amount")
			return cc.app.Course([]string{"budget", "distribute", args[0], "--amount", amount})
		},
	}
	distributeCmd.Flags().String("amount", "", "Budget per student in USD (required)")
	_ = distributeCmd.MarkFlagRequired("amount")

	cmd.AddCommand(showCmd, distributeCmd)
	return cmd
}

// TACobraCommands handles TA operations.
type TACobraCommands struct {
	app *App
}

// NewTACobraCommands creates new TA cobra commands.
func NewTACobraCommands(app *App) *TACobraCommands {
	return &TACobraCommands{app: app}
}

// CreateTACommand creates the 'ta' top-level command.
func (tc *TACobraCommands) CreateTACommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ta",
		Short: "TA and instructor operations for course management",
		Long: `Instructor/TA tools for debugging and resetting student environments.

Examples:
  prism ta debug <course-id> <student-id>
  prism ta reset <course-id> <student-id> --reason "broken environment"`,
	}
	cmd.AddCommand(tc.createDebugCommand(), tc.createResetCommand())
	return cmd
}

func (tc *TACobraCommands) createDebugCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "debug <course-id> <student-id>",
		Short: "View student environment status and recent events",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return tc.app.TA([]string{"debug", args[0], args[1]})
		},
	}
}

func (tc *TACobraCommands) createResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset <course-id> <student-id>",
		Short: "Reset a student's instance environment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			reason, _ := cmd.Flags().GetString("reason")
			retention, _ := cmd.Flags().GetString("retention")

			cliArgs := []string{"reset", args[0], args[1]}
			if reason != "" {
				cliArgs = append(cliArgs, "--reason", reason)
			}
			if retention != "" {
				cliArgs = append(cliArgs, "--retention", retention)
			}
			return tc.app.TA(cliArgs)
		},
	}
	cmd.Flags().String("reason", "", "Reason for reset (required for audit trail)")
	cmd.Flags().String("retention", "7", "Backup retention in days")
	_ = cmd.MarkFlagRequired("reason")
	return cmd
}
