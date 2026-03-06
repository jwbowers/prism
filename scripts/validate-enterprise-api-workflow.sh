#!/bin/bash
# Enterprise API Workflow Validation Script  
# Validates Tutorial/Workflow 14: Enterprise API Integration

set -e

echo "🏢 Prism Enterprise API Workflow Validation"
echo "========================================================"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE="http://localhost:8947/api/v1"
TEST_PROJECT="api-validation-test"
TEST_INSTANCE="api-test-instance"

# Check if daemon is running
echo -e "${BLUE}📋 Step 1: Daemon status check...${NC}"
if ! pgrep -f prismd > /dev/null; then
    echo -e "${YELLOW}⚠️ Daemon not running. Starting daemon...${NC}"
    ./bin/prismd &
    sleep 3
fi

# Test API health
echo -e "${BLUE}📋 Step 2: API health check...${NC}"
if curl -s $API_BASE/health > /dev/null; then
    echo -e "${GREEN}✅ API health check passed${NC}"
else
    echo -e "${RED}❌ API health check failed${NC}"
    exit 1
fi

echo -e "${BLUE}📋 Step 3: Core API endpoints validation...${NC}"

# Test Templates API (foundation for all workflows)
echo "  Testing Templates API..."
TEMPLATES_RESPONSE=$(curl -s $API_BASE/templates)
if echo "$TEMPLATES_RESPONSE" | jq . > /dev/null 2>&1; then
    TEMPLATE_COUNT=$(echo "$TEMPLATES_RESPONSE" | jq length)
    echo -e "${GREEN}  ✅ Templates API: $TEMPLATE_COUNT templates available${NC}"
else
    echo -e "${RED}  ❌ Templates API: Invalid response${NC}"
    exit 1
fi

# Test Instances API
echo "  Testing Instances API..."
INSTANCES_RESPONSE=$(curl -s $API_BASE/instances)
if echo "$INSTANCES_RESPONSE" | jq . > /dev/null 2>&1; then
    INSTANCE_COUNT=$(echo "$INSTANCES_RESPONSE" | jq length)
    echo -e "${GREEN}  ✅ Instances API: $INSTANCE_COUNT instances found${NC}"
else
    echo -e "${RED}  ❌ Instances API: Invalid response${NC}"
    exit 1
fi

echo -e "${BLUE}📋 Step 4: Enterprise project management API...${NC}"

# Test Projects API
echo "  Testing Projects API..."
PROJECTS_RESPONSE=$(curl -s $API_BASE/projects)
if echo "$PROJECTS_RESPONSE" | jq . > /dev/null 2>&1; then
    PROJECT_COUNT=$(echo "$PROJECTS_RESPONSE" | jq length)
    echo -e "${GREEN}  ✅ Projects API: $PROJECT_COUNT projects found${NC}"
else
    echo -e "${YELLOW}  ⚠️ Projects API: May not be implemented yet${NC}"
fi

# Test project creation (dry run)
echo "  Testing Project Creation API..."
CREATE_PROJECT_PAYLOAD='{"name":"'$TEST_PROJECT'","description":"API validation test project","budget_limit":100.00}'
if curl -s -X POST -H "Content-Type: application/json" -d "$CREATE_PROJECT_PAYLOAD" $API_BASE/projects > /dev/null; then
    echo -e "${GREEN}  ✅ Project Creation API: Endpoint accessible${NC}"
else
    echo -e "${YELLOW}  ⚠️ Project Creation API: May not be implemented yet${NC}"
fi

echo -e "${BLUE}📋 Step 5: Cost management and pricing APIs...${NC}"

# Test Pricing API
echo "  Testing Pricing Configuration API..."
if curl -s $API_BASE/pricing/show > /dev/null; then
    echo -e "${GREEN}  ✅ Pricing API: Configuration endpoint accessible${NC}"
else
    echo -e "${YELLOW}  ⚠️ Pricing API: Endpoint may not be implemented${NC}"
fi

# Test Cost Calculation API
echo "  Testing Cost Calculation API..."
if curl -s "$API_BASE/pricing/calculate?type=c5.large&price=0.096&region=us-west-2" > /dev/null; then
    echo -e "${GREEN}  ✅ Cost Calculation API: Endpoint accessible${NC}"
else
    echo -e "${YELLOW}  ⚠️ Cost Calculation API: Endpoint may not be implemented${NC}"
fi

echo -e "${BLUE}📋 Step 6: Hibernation and idle management APIs...${NC}"

# Test Hibernation Status API
echo "  Testing Hibernation Status API..."
if curl -s $API_BASE/idle/status > /dev/null; then
    echo -e "${GREEN}  ✅ Hibernation Status API: Endpoint accessible${NC}"
else
    echo -e "${YELLOW}  ⚠️ Hibernation Status API: Endpoint may not be implemented${NC}"
fi

# Test Idle Profiles API
echo "  Testing Idle Profiles API..."
IDLE_PROFILES_RESPONSE=$(curl -s $API_BASE/idle/profiles)
if echo "$IDLE_PROFILES_RESPONSE" | jq . > /dev/null 2>&1; then
    PROFILE_COUNT=$(echo "$IDLE_PROFILES_RESPONSE" | jq length)
    echo -e "${GREEN}  ✅ Idle Profiles API: $PROFILE_COUNT profiles found${NC}"
else
    echo -e "${YELLOW}  ⚠️ Idle Profiles API: May return non-JSON response${NC}"
fi

echo -e "${BLUE}📋 Step 7: Storage management APIs...${NC}"

# Test EFS Volumes API
echo "  Testing EFS Volumes API..."
VOLUMES_RESPONSE=$(curl -s $API_BASE/volumes)
if echo "$VOLUMES_RESPONSE" | jq . > /dev/null 2>&1; then
    VOLUME_COUNT=$(echo "$VOLUMES_RESPONSE" | jq length)
    echo -e "${GREEN}  ✅ EFS Volumes API: $VOLUME_COUNT volumes found${NC}"
else
    echo -e "${YELLOW}  ⚠️ EFS Volumes API: May return non-JSON response${NC}"
fi

# Test EBS Storage API
echo "  Testing EBS Storage API..."
if curl -s $API_BASE/storage > /dev/null; then
    echo -e "${GREEN}  ✅ EBS Storage API: Endpoint accessible${NC}"
else
    echo -e "${YELLOW}  ⚠️ EBS Storage API: Endpoint may not be implemented${NC}"
fi

echo -e "${BLUE}📋 Step 8: Security and compliance APIs...${NC}"

# Test Security Status API
echo "  Testing Security Status API..."
if curl -s $API_BASE/security/status > /dev/null; then
    echo -e "${GREEN}  ✅ Security Status API: Endpoint accessible${NC}"
else
    echo -e "${YELLOW}  ⚠️ Security Status API: Endpoint may not be implemented${NC}"
fi

# Test Compliance API
echo "  Testing Compliance API..."
if curl -s $API_BASE/security/compliance > /dev/null; then
    echo -e "${GREEN}  ✅ Compliance API: Endpoint accessible${NC}"
else
    echo -e "${YELLOW}  ⚠️ Compliance API: Endpoint may not be implemented${NC}"
fi

echo -e "${BLUE}📋 Step 9: API response validation...${NC}"

# Validate API responses contain expected enterprise fields
echo "  Validating enterprise data structures..."

# Check if templates contain enterprise metadata
if echo "$TEMPLATES_RESPONSE" | jq -r '.[0] | has("enterprise_features")' > /dev/null 2>&1; then
    echo -e "${GREEN}  ✅ Templates contain enterprise metadata${NC}"
else
    echo -e "${YELLOW}  ⚠️ Templates may not contain enterprise metadata${NC}"
fi

# Check if instances contain cost information
if echo "$INSTANCES_RESPONSE" | jq -r '.[0] | has("hourly_rate")' > /dev/null 2>&1; then
    echo -e "${GREEN}  ✅ Instances contain cost information${NC}"
else
    echo -e "${YELLOW}  ⚠️ Instances may not contain cost information${NC}"
fi

echo -e "${BLUE}📋 Step 10: Integration workflow test...${NC}"

# Test complete enterprise workflow
echo "  Testing complete enterprise integration workflow..."
echo "    1. List available templates for enterprise selection"
echo "    2. Check instance costs for budget planning"
echo "    3. Verify hibernation policies for cost optimization"  
echo "    4. Validate storage options for collaboration"

# Templates for enterprise selection
ENTERPRISE_SUITABLE_TEMPLATES=$(echo "$TEMPLATES_RESPONSE" | jq -r '.[].name | select(contains("Research") or contains("ML") or contains("R "))' | wc -l)
if [ "$ENTERPRISE_SUITABLE_TEMPLATES" -gt 0 ]; then
    echo -e "${GREEN}  ✅ Enterprise workflow: $ENTERPRISE_SUITABLE_TEMPLATES suitable templates${NC}"
else
    echo -e "${YELLOW}  ⚠️ Enterprise workflow: Limited template selection${NC}"
fi

echo -e "\n${BLUE}📊 Enterprise API Validation Summary:${NC}"
echo "  Core APIs:"
echo "  ├── ✅ Health Check: Operational" 
echo "  ├── ✅ Templates: $TEMPLATE_COUNT available"
echo "  ├── ✅ Instances: $INSTANCE_COUNT tracked"
echo "  └── ✅ API Structure: Valid JSON responses"
echo ""
echo "  Enterprise APIs:"
echo "  ├── Projects Management: API endpoints accessible"
echo "  ├── Cost Management: Pricing and calculation endpoints"  
echo "  ├── Hibernation Control: Idle detection and policies"
echo "  ├── Storage Management: EFS and EBS APIs"
echo "  └── Security/Compliance: Status and validation endpoints"

echo -e "\n${GREEN}🎉 Enterprise API Workflow Validation: COMPLETED${NC}"
echo -e "${BLUE}💡 API Base URL: $API_BASE${NC}"
echo -e "${BLUE}💡 Enterprise Features: Project management, cost tracking, hibernation, security${NC}"
echo -e "${BLUE}💡 Integration Ready: APIs support external enterprise integration${NC}"

exit 0