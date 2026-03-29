package cli

import (
	"fmt"

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
		cc.createArchiveCommand(),
		cc.createReportCommand(),
		cc.createAuditCommand(),
		cc.createTAAccessCommand(),
		cc.createMaterialsCommand(),
		cc.createResetWorkspaceCommand(),
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

	importCmd := &cobra.Command{
		Use:   "import <course-id> <file>",
		Short: "Import a student roster from CSV (Prism/Canvas/Blackboard)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			cliArgs := []string{"students", "import", args[0], args[1]}
			if format != "" {
				cliArgs = append(cliArgs, "--format", format)
			}
			return cc.app.Course(cliArgs)
		},
	}
	importCmd.Flags().String("format", "prism", "Roster format: prism, canvas, blackboard")

	provisionCmd := &cobra.Command{
		Use:   "provision <course-id> <user-id>",
		Short: "Provision a workspace for a student",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			template, _ := cmd.Flags().GetString("template")
			name, _ := cmd.Flags().GetString("name")
			cliArgs := []string{"students", "provision", args[0], args[1]}
			if template != "" {
				cliArgs = append(cliArgs, "--template", template)
			}
			if name != "" {
				cliArgs = append(cliArgs, "--name", name)
			}
			return cc.app.Course(cliArgs)
		},
	}
	provisionCmd.Flags().String("template", "", "Template slug override")
	provisionCmd.Flags().String("name", "", "Instance name override")

	cmd.AddCommand(listCmd, enrollCmd, unenrollCmd, importCmd, provisionCmd)
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
	cmd.AddCommand(tc.createDebugCommand(), tc.createResetCommand(), tc.createOverviewCommand())
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

func (tc *TACobraCommands) createOverviewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "overview <course-id>",
		Short: "Show TA dashboard overview for a course",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return tc.app.TA([]string{"overview", args[0]})
		},
	}
}

// v0.16.0 course commands

func (cc *CourseCobraCommands) createArchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <course-id>",
		Short: "Archive a course and stop its instances",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"archive", args[0]})
		},
	}
}

func (cc *CourseCobraCommands) createReportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report <course-id>",
		Short: "Generate a student usage report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			output, _ := cmd.Flags().GetString("output")
			cliArgs := []string{"report", args[0]}
			if format != "" {
				cliArgs = append(cliArgs, "--format", format)
			}
			if output != "" {
				cliArgs = append(cliArgs, "--output", output)
			}
			return cc.app.Course(cliArgs)
		},
	}
	cmd.Flags().String("format", "", "Output format: json or csv")
	cmd.Flags().String("output", "", "Write output to file")
	return cmd
}

func (cc *CourseCobraCommands) createAuditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit <course-id>",
		Short: "Query the academic integrity audit log",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			student, _ := cmd.Flags().GetString("student")
			since, _ := cmd.Flags().GetString("since")
			limitStr, _ := cmd.Flags().GetString("limit")
			cliArgs := []string{"audit", args[0]}
			if student != "" {
				cliArgs = append(cliArgs, "--student", student)
			}
			if since != "" {
				cliArgs = append(cliArgs, "--since", since)
			}
			if limitStr != "" {
				cliArgs = append(cliArgs, "--limit", limitStr)
			}
			return cc.app.Course(cliArgs)
		},
	}
	cmd.Flags().String("student", "", "Filter by student user ID or email")
	cmd.Flags().String("since", "", "Filter entries since this date (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().String("limit", "100", "Maximum number of entries to return")
	return cmd
}

// --- v0.19.0: TA Access, Materials, Reset Workspace ---

// createTAAccessCommand creates the 'course ta-access' command group.
func (cc *CourseCobraCommands) createTAAccessCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ta-access",
		Short: "Manage TA access to student instances",
		Long: `Grant, revoke, and audit TA access to student workspaces.

Examples:
  prism course ta-access list <course-id>
  prism course ta-access grant <course-id> --email ta@uni.edu
  prism course ta-access revoke <course-id> --email ta@uni.edu
  prism course ta-access connect <course-id> --student student@uni.edu --reason "office hours"`,
	}

	// ta-access list
	listCmd := &cobra.Command{
		Use:   "list <course-id>",
		Short: "List TAs with access",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"ta-access", "list", args[0]})
		},
	}

	// ta-access grant
	grantCmd := &cobra.Command{
		Use:   "grant <course-id>",
		Short: "Grant TA access to a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			displayName, _ := cmd.Flags().GetString("name")
			if email == "" {
				return cmd.Usage()
			}
			cliArgs := []string{"ta-access", "grant", args[0], "--email", email}
			if displayName != "" {
				cliArgs = append(cliArgs, "--name", displayName)
			}
			return cc.app.Course(cliArgs)
		},
	}
	grantCmd.Flags().String("email", "", "TA email address (required)")
	grantCmd.Flags().String("name", "", "TA display name")
	_ = grantCmd.MarkFlagRequired("email")

	// ta-access revoke
	revokeCmd := &cobra.Command{
		Use:   "revoke <course-id>",
		Short: "Revoke TA access from a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			if email == "" {
				return cmd.Usage()
			}
			return cc.app.Course([]string{"ta-access", "revoke", args[0], "--email", email})
		},
	}
	revokeCmd.Flags().String("email", "", "TA email address (required)")

	// ta-access connect
	connectCmd := &cobra.Command{
		Use:   "connect <course-id>",
		Short: "Get SSH command to connect to a student's instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			student, _ := cmd.Flags().GetString("student")
			reason, _ := cmd.Flags().GetString("reason")
			if student == "" || reason == "" {
				return cmd.Usage()
			}
			return cc.app.Course([]string{"ta-access", "connect", args[0], "--student", student, "--reason", reason})
		},
	}
	connectCmd.Flags().String("student", "", "Student user ID or email (required)")
	connectCmd.Flags().String("reason", "", "Reason for access — recorded in audit log (required)")

	cmd.AddCommand(listCmd, grantCmd, revokeCmd, connectCmd)
	return cmd
}

// createMaterialsCommand creates the 'course materials' command group.
func (cc *CourseCobraCommands) createMaterialsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "materials",
		Short: "Manage shared course materials (read-only EFS)",
		Long: `Create and manage a shared EFS volume for course datasets and assignments.

Examples:
  prism course materials create <course-id> --size 50 --mount /mnt/course-materials
  prism course materials list <course-id>
  prism course materials mount <course-id>`,
	}

	// materials create
	createCmd := &cobra.Command{
		Use:   "create <course-id>",
		Short: "Create shared EFS materials volume",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			size, _ := cmd.Flags().GetInt("size")
			mount, _ := cmd.Flags().GetString("mount")
			cliArgs := []string{"materials", "create", args[0], "--size", fmt.Sprintf("%d", size)}
			if mount != "" {
				cliArgs = append(cliArgs, "--mount", mount)
			}
			return cc.app.Course(cliArgs)
		},
	}
	createCmd.Flags().Int("size", 50, "Advisory size in GB (EFS is elastic)")
	createCmd.Flags().String("mount", "/mnt/course-materials", "Mount path in student instances")

	// materials list
	listCmd := &cobra.Command{
		Use:   "list <course-id>",
		Short: "Show materials volume info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"materials", "list", args[0]})
		},
	}

	// materials mount
	mountCmd := &cobra.Command{
		Use:   "mount <course-id>",
		Short: "Schedule EFS mount on all student instances",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.app.Course([]string{"materials", "mount", args[0]})
		},
	}

	cmd.AddCommand(createCmd, listCmd, mountCmd)
	return cmd
}

// createResetWorkspaceCommand creates 'course reset-workspace'.
func (cc *CourseCobraCommands) createResetWorkspaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset-workspace <course-id>",
		Short: "Reset a student's workspace (TA/instructor only)",
		Long: `Resets a student's workspace to a clean state. Optionally creates a backup
snapshot before resetting so the student can retrieve their work.

Examples:
  prism course reset-workspace CS229 --student student@uni.edu --reason "broken env" --backup
  prism course reset-workspace CS229 --student student@uni.edu --reason "fresh start"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			student, _ := cmd.Flags().GetString("student")
			reason, _ := cmd.Flags().GetString("reason")
			backup, _ := cmd.Flags().GetBool("backup")
			if student == "" || reason == "" {
				return cmd.Usage()
			}
			cliArgs := []string{"reset-workspace", args[0], "--student", student, "--reason", reason}
			if backup {
				cliArgs = append(cliArgs, "--backup")
			}
			return cc.app.Course(cliArgs)
		},
	}
	cmd.Flags().String("student", "", "Student user ID or email (required)")
	cmd.Flags().String("reason", "", "Reason for reset (required, recorded in audit log)")
	cmd.Flags().Bool("backup", false, "Create a backup snapshot before resetting")
	_ = cmd.MarkFlagRequired("student")
	return cmd
}
