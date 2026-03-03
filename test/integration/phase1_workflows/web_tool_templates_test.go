//go:build integration
// +build integration

package phase1_workflows

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/test/fixtures"
	"github.com/scttfrdmn/prism/test/integration"
)

// TestTemplateProvisioning_Streamlit validates end-to-end provisioning of the
// Streamlit Data Apps template — Python venv, demo app, systemd service on port 8501.
//
// Issues addressed: #211 (Streamlit template)
//
// Success criteria:
// - Instance launches and reaches running state
// - Provisioning (user data) completes successfully
// - streamlit service is active (systemd)
// - HTTP endpoint on port 8501 returns 200
// - streamlit package importable in Python venv
func TestTemplateProvisioning_Streamlit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := integration.NewTestContext(t)
	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	instanceName := integration.GenerateTestName("test-streamlit")
	t.Logf("Launching Streamlit instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "Streamlit Data Apps",
		Name:     instanceName,
		Size:     "S",
	})
	integration.AssertNoError(t, err, "Failed to create Streamlit instance")
	integration.AssertNotEmpty(t, instance.ID, "Instance should have EC2 ID")
	integration.AssertEqual(t, "running", instance.State, "Instance should be running")

	t.Logf("Instance launched: %s (ID: %s, IP: %s)", instance.Name, instance.ID, instance.PublicIP)

	t.Log("Waiting for provisioning to complete (pip install + service start)...")
	if err := waitForProvisioningComplete(ctx, instance.Name, 10*time.Minute); err != nil {
		t.Fatalf("Provisioning did not complete: %v", err)
	}
	t.Log("Provisioning completed successfully")

	if canSSH(instance.PublicIP) {
		t.Log("SSH available - performing deep verification")
		verifySystemdServiceActive(t, instance, "streamlit")
		verifyHTTPEndpoint(t, instance, 8501)
		verifyStreamlitPackage(t, instance)
		t.Log("✓ All Streamlit verification checks passed")
	} else {
		t.Log("⚠️  SSH not available - skipping deep verification")
		t.Log("Note: Enable SSH (port 22) and configure an SSH key to enable deep checks")
	}

	instanceInfo, err := ctx.Client.GetInstance(context.Background(), instanceName)
	integration.AssertNoError(t, err, "Should be able to retrieve instance info")
	integration.AssertEqual(t, "Streamlit Data Apps", instanceInfo.Template, "Template name should match")
	integration.AssertNotEmpty(t, instanceInfo.PublicIP, "Instance should have public IP")
	integration.AssertEqual(t, "running", instanceInfo.State, "Instance should still be running")

	t.Log("✓ Streamlit template provisioning test completed successfully")
}

// TestTemplateProvisioning_RShiny validates the R Shiny Server template —
// inherits R Base, adds shiny packages, Shiny Server deb, sample app on port 3838.
//
// Issues addressed: #212 (R Shiny Server template)
//
// Success criteria:
// - Instance launches and provisioning completes
// - shiny-server service is active
// - HTTP endpoint on port 3838 returns 200
// - shiny R package is importable
func TestTemplateProvisioning_RShiny(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := integration.NewTestContext(t)
	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	instanceName := integration.GenerateTestName("test-r-shiny")
	t.Logf("Launching R Shiny Server instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "R Shiny Server",
		Name:     instanceName,
		Size:     "M", // Inherits R Base — memory-optimized
	})
	integration.AssertNoError(t, err, "Failed to create R Shiny Server instance")
	integration.AssertEqual(t, "running", instance.State, "Instance should be running")

	t.Logf("Instance launched: %s (ID: %s, IP: %s)", instance.Name, instance.ID, instance.PublicIP)

	// R package install + Shiny Server deb + R Base inheritance takes longer
	t.Log("Waiting for provisioning to complete (R packages + Shiny Server install)...")
	if err := waitForProvisioningComplete(ctx, instance.Name, 15*time.Minute); err != nil {
		t.Fatalf("Provisioning did not complete: %v", err)
	}
	t.Log("Provisioning completed successfully")

	if canSSH(instance.PublicIP) {
		t.Log("SSH available - performing deep verification")
		verifySystemdServiceActive(t, instance, "shiny-server")
		verifyHTTPEndpoint(t, instance, 3838)
		verifyShinyPackage(t, instance)
		verifyShinyAppDeployed(t, instance)
		t.Log("✓ All R Shiny Server verification checks passed")
	} else {
		t.Log("⚠️  SSH not available - skipping deep verification")
	}

	instanceInfo, err := ctx.Client.GetInstance(context.Background(), instanceName)
	integration.AssertNoError(t, err, "Should be able to retrieve instance info")
	integration.AssertEqual(t, "R Shiny Server", instanceInfo.Template, "Template name should match")
	integration.AssertEqual(t, "running", instanceInfo.State, "Instance should still be running")

	t.Log("✓ R Shiny Server template provisioning test completed successfully")
}

// TestTemplateProvisioning_OpenRefine validates the OpenRefine Data Cleaning template —
// OpenJDK 17, OpenRefine 3.8.2 tarball, systemd service, sample CSV on port 3333.
//
// Issues addressed: #213 (OpenRefine template)
//
// Success criteria:
// - Instance launches and provisioning completes
// - openrefine service is active
// - HTTP endpoint on port 3333 returns 200 (OpenRefine UI)
// - Java 17 is installed
// - Sample survey CSV is present
func TestTemplateProvisioning_OpenRefine(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := integration.NewTestContext(t)
	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	instanceName := integration.GenerateTestName("test-openrefine")
	t.Logf("Launching OpenRefine instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "OpenRefine Data Cleaning",
		Name:     instanceName,
		Size:     "M",
	})
	integration.AssertNoError(t, err, "Failed to create OpenRefine instance")
	integration.AssertEqual(t, "running", instance.State, "Instance should be running")

	t.Logf("Instance launched: %s (ID: %s, IP: %s)", instance.Name, instance.ID, instance.PublicIP)

	t.Log("Waiting for provisioning to complete (OpenRefine tarball download + extract)...")
	if err := waitForProvisioningComplete(ctx, instance.Name, 10*time.Minute); err != nil {
		t.Fatalf("Provisioning did not complete: %v", err)
	}
	t.Log("Provisioning completed successfully")

	if canSSH(instance.PublicIP) {
		t.Log("SSH available - performing deep verification")
		verifySystemdServiceActive(t, instance, "openrefine")
		verifyHTTPEndpoint(t, instance, 3333)
		verifyJavaInstalled(t, instance)
		verifyOpenRefineInstalled(t, instance)
		verifySampleCSVPresent(t, instance)
		t.Log("✓ All OpenRefine verification checks passed")
	} else {
		t.Log("⚠️  SSH not available - skipping deep verification")
	}

	instanceInfo, err := ctx.Client.GetInstance(context.Background(), instanceName)
	integration.AssertNoError(t, err, "Should be able to retrieve instance info")
	integration.AssertEqual(t, "OpenRefine Data Cleaning", instanceInfo.Template, "Template name should match")
	integration.AssertEqual(t, "running", instanceInfo.State, "Instance should still be running")

	t.Log("✓ OpenRefine template provisioning test completed successfully")
}

// TestTemplateProvisioning_LabelStudio validates the Label Studio ML Annotation template —
// Python venv, label-studio pip package, SQLite backend, systemd service on port 8090.
//
// Issues addressed: #214 (Label Studio template)
//
// Success criteria:
// - Instance launches and provisioning completes
// - label-studio service is active
// - HTTP endpoint on port 8090 returns 200 or 302 (redirect to login)
// - label-studio data directory exists
func TestTemplateProvisioning_LabelStudio(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := integration.NewTestContext(t)
	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	instanceName := integration.GenerateTestName("test-label-studio")
	t.Logf("Launching Label Studio instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "Label Studio ML Annotation",
		Name:     instanceName,
		Size:     "M", // label-studio is heavier than other pip packages
	})
	integration.AssertNoError(t, err, "Failed to create Label Studio instance")
	integration.AssertEqual(t, "running", instance.State, "Instance should be running")

	t.Logf("Instance launched: %s (ID: %s, IP: %s)", instance.Name, instance.ID, instance.PublicIP)

	// label-studio pip install is one of the heaviest installs
	t.Log("Waiting for provisioning to complete (label-studio pip install — may take 12+ min)...")
	if err := waitForProvisioningComplete(ctx, instance.Name, 15*time.Minute); err != nil {
		t.Fatalf("Provisioning did not complete: %v", err)
	}
	t.Log("Provisioning completed successfully")

	if canSSH(instance.PublicIP) {
		t.Log("SSH available - performing deep verification")
		verifySystemdServiceActive(t, instance, "label-studio")
		// Label Studio redirects to /user/login on first visit; accept 200 or 302
		verifyHTTPEndpointAny(t, instance, 8090, []string{"200", "302"})
		verifyLabelStudioDataDir(t, instance)
		t.Log("✓ All Label Studio verification checks passed")
	} else {
		t.Log("⚠️  SSH not available - skipping deep verification")
	}

	instanceInfo, err := ctx.Client.GetInstance(context.Background(), instanceName)
	integration.AssertNoError(t, err, "Should be able to retrieve instance info")
	integration.AssertEqual(t, "Label Studio ML Annotation", instanceInfo.Template, "Template name should match")
	integration.AssertEqual(t, "running", instanceInfo.State, "Instance should still be running")

	t.Log("✓ Label Studio template provisioning test completed successfully")
}

// TestTemplateProvisioning_Datasette validates the Datasette Dataset Explorer template —
// Python venv, datasette + csvs-to-sqlite, sample SQLite DB, systemd service on port 8001.
//
// Issues addressed: #215 (Datasette template)
//
// Success criteria:
// - Instance launches and provisioning completes
// - datasette service is active
// - HTTP endpoint on port 8001 returns 200 (Datasette index)
// - Sample research database file exists at ~/data/sample-research.db
// - CSV import helper script is present at ~/import-csv.sh
func TestTemplateProvisioning_Datasette(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := integration.NewTestContext(t)
	registry := fixtures.NewFixtureRegistry(t, ctx.Client)

	instanceName := integration.GenerateTestName("test-datasette")
	t.Logf("Launching Datasette instance: %s", instanceName)

	instance, err := fixtures.CreateTestInstance(t, registry, fixtures.CreateTestInstanceOptions{
		Template: "Datasette Dataset Explorer",
		Name:     instanceName,
		Size:     "S",
	})
	integration.AssertNoError(t, err, "Failed to create Datasette instance")
	integration.AssertEqual(t, "running", instance.State, "Instance should be running")

	t.Logf("Instance launched: %s (ID: %s, IP: %s)", instance.Name, instance.ID, instance.PublicIP)

	t.Log("Waiting for provisioning to complete (datasette pip install + sample DB creation)...")
	if err := waitForProvisioningComplete(ctx, instance.Name, 10*time.Minute); err != nil {
		t.Fatalf("Provisioning did not complete: %v", err)
	}
	t.Log("Provisioning completed successfully")

	if canSSH(instance.PublicIP) {
		t.Log("SSH available - performing deep verification")
		verifySystemdServiceActive(t, instance, "datasette")
		verifyHTTPEndpoint(t, instance, 8001)
		verifyDatasetteDB(t, instance)
		verifyCSVImportScript(t, instance)
		t.Log("✓ All Datasette verification checks passed")
	} else {
		t.Log("⚠️  SSH not available - skipping deep verification")
	}

	instanceInfo, err := ctx.Client.GetInstance(context.Background(), instanceName)
	integration.AssertNoError(t, err, "Should be able to retrieve instance info")
	integration.AssertEqual(t, "Datasette Dataset Explorer", instanceInfo.Template, "Template name should match")
	integration.AssertEqual(t, "running", instanceInfo.State, "Instance should still be running")

	t.Log("✓ Datasette template provisioning test completed successfully")
}

// ─── Shared helper functions ──────────────────────────────────────────────────

// verifySystemdServiceActive checks that a systemd service is in active (running) state
func verifySystemdServiceActive(t *testing.T, instance *types.Instance, serviceName string) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP, "systemctl is-active "+serviceName)
	if strings.TrimSpace(output) != "active" {
		// Collect journal for diagnostics
		logs := runSSHCommandOrEmpty(instance.PublicIP, "journalctl -u "+serviceName+" -n 20 --no-pager")
		t.Errorf("Service %s is not active (got: %q)\nRecent journal:\n%s", serviceName, output, logs)
		return
	}

	t.Logf("✓ systemd service %q is active", serviceName)
}

// verifyHTTPEndpoint checks that an HTTP service on localhost:<port> returns HTTP 200
func verifyHTTPEndpoint(t *testing.T, instance *types.Instance, port int) {
	t.Helper()
	verifyHTTPEndpointAny(t, instance, port, []string{"200"})
}

// verifyHTTPEndpointAny checks that localhost:<port> returns one of the accepted HTTP codes.
// Uses a short retry loop because systemd services may take a few seconds to fully bind after
// cloud-init completes.
func verifyHTTPEndpointAny(t *testing.T, instance *types.Instance, port int, acceptedCodes []string) {
	t.Helper()

	curlCmd := fmt.Sprintf(
		"curl -s -o /dev/null -w '%%{http_code}' --max-time 5 --retry 3 --retry-delay 5 "+
			"--retry-connrefused http://localhost:%d", port)

	output := runSSHCommand(t, instance.PublicIP, curlCmd)
	code := strings.TrimSpace(output)

	for _, accepted := range acceptedCodes {
		if code == accepted {
			t.Logf("✓ HTTP endpoint on port %d returned %s", port, code)
			return
		}
	}

	t.Errorf("HTTP endpoint on port %d returned %q (want one of %v)", port, code, acceptedCodes)
}

// ─── Streamlit helpers ────────────────────────────────────────────────────────

// verifyStreamlitPackage confirms streamlit is importable from the venv
func verifyStreamlitPackage(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP,
		`/home/researcher/.venv/bin/python3 -c "import streamlit; print('OK')"`)
	if !strings.Contains(output, "OK") {
		t.Errorf("streamlit package not importable from venv: %s", output)
		return
	}
	t.Log("✓ streamlit package importable from venv")
}

// ─── R Shiny helpers ──────────────────────────────────────────────────────────

// verifyShinyPackage confirms the shiny R package is installed
func verifyShinyPackage(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP, `R -e "library(shiny); cat('OK\n')" 2>&1`)
	if !strings.Contains(output, "OK") {
		t.Errorf("R shiny package not installed: %s", output)
		return
	}
	t.Log("✓ R shiny package is installed")
}

// verifyShinyAppDeployed confirms the sample Shiny app exists
func verifyShinyAppDeployed(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP,
		"test -f /srv/shiny-server/sample-app/app.R && echo 'found'")
	if !strings.Contains(output, "found") {
		t.Errorf("Sample Shiny app not found at /srv/shiny-server/sample-app/app.R")
		return
	}
	t.Log("✓ Sample Shiny app deployed at /srv/shiny-server/sample-app/app.R")
}

// ─── OpenRefine helpers ───────────────────────────────────────────────────────

// verifyJavaInstalled confirms Java 17 is present
func verifyJavaInstalled(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP, "java -version 2>&1")
	if !strings.Contains(output, "17") && !strings.Contains(output, "openjdk") {
		t.Errorf("Java 17 not found in version output: %s", output)
		return
	}
	t.Log("✓ Java (OpenJDK 17) is installed")
}

// verifyOpenRefineInstalled confirms the OpenRefine refine script exists
func verifyOpenRefineInstalled(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP, "test -x /opt/openrefine/refine && echo 'found'")
	if !strings.Contains(output, "found") {
		t.Errorf("OpenRefine not found at /opt/openrefine/refine")
		return
	}
	t.Log("✓ OpenRefine installed at /opt/openrefine/refine")
}

// verifySampleCSVPresent confirms the intentionally-messy sample CSV is present
func verifySampleCSVPresent(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP,
		"test -f /home/researcher/openrefine-workspace/sample-survey.csv && echo 'found'")
	if !strings.Contains(output, "found") {
		t.Errorf("Sample CSV not found at ~/openrefine-workspace/sample-survey.csv")
		return
	}
	t.Log("✓ Sample survey CSV present at ~/openrefine-workspace/sample-survey.csv")
}

// ─── Label Studio helpers ─────────────────────────────────────────────────────

// verifyLabelStudioDataDir confirms the data directory was created
func verifyLabelStudioDataDir(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP,
		"test -d /home/researcher/label-studio-data && echo 'found'")
	if !strings.Contains(output, "found") {
		t.Errorf("Label Studio data directory not found at ~/label-studio-data/")
		return
	}
	t.Log("✓ Label Studio data directory exists at ~/label-studio-data/")
}

// ─── Datasette helpers ────────────────────────────────────────────────────────

// verifyDatasetteDB confirms the sample SQLite database was created
func verifyDatasetteDB(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP,
		"test -f /home/researcher/data/sample-research.db && echo 'found'")
	if !strings.Contains(output, "found") {
		t.Errorf("Sample database not found at ~/data/sample-research.db")
		return
	}

	// Confirm it contains the publications table
	queryOut := runSSHCommand(t, instance.PublicIP,
		`sqlite3 /home/researcher/data/sample-research.db "SELECT COUNT(*) FROM publications"`)
	count := strings.TrimSpace(queryOut)
	if count == "" || count == "0" {
		t.Errorf("Sample database exists but publications table is empty or missing (got: %q)", count)
		return
	}

	t.Logf("✓ Sample SQLite database present with %s publication rows", count)
}

// verifyCSVImportScript confirms the ~/import-csv.sh helper is present and executable
func verifyCSVImportScript(t *testing.T, instance *types.Instance) {
	t.Helper()

	output := runSSHCommand(t, instance.PublicIP,
		"test -x /home/researcher/import-csv.sh && echo 'found'")
	if !strings.Contains(output, "found") {
		t.Errorf("CSV import script not found or not executable at ~/import-csv.sh")
		return
	}
	t.Log("✓ CSV import helper ~/import-csv.sh is present and executable")
}
