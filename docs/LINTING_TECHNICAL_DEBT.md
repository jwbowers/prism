# Linting Technical Debt - v0.5.16

This document tracks known linting issues that are documented but not blocking for the v0.5.16 release.

## Summary

- **Go Cyclomatic Complexity**: 40 functions with complexity > 15
- **TypeScript/React ESLint**: 97 problems (37 errors, 60 warnings)

## Go Cyclomatic Complexity Issues

### Threshold: Functions with complexity > 15

**Total**: 40 functions

#### Severity Distribution
- **High Complexity (30+)**: 3 functions
- **Medium Complexity (20-29)**: 9 functions
- **Low Complexity (16-19)**: 28 functions

### Top 10 Most Complex Functions

| Complexity | Function | File | Line |
|------------|----------|------|------|
| 33 | `(*App).projectInvitationsBulk` | internal/cli/invitation_impl.go | 783 |
| 30 | `(*App).List` | internal/cli/app.go | 618 |
| 29 | `(PolicyModel).Update` | internal/tui/models/policy.go | 105 |
| 28 | `(*InstanceBuilder).BuildInstance` | pkg/aws/manager.go | 2257 |
| 26 | `(*Server).handleInvitationOperations` | pkg/daemon/invitation_handlers.go | 20 |
| 24 | `DefaultTemplateDirs` | pkg/templates/templates.go | 20 |
| 24 | `testProfileOperations` | pkg/profile/enhanced_manager_functional_test.go | 40 |
| 24 | `(*Server).handleLaunchInstance` | pkg/daemon/instance_handlers.go | 150 |
| 23 | `(*PageNavigationHandler).Handle` | internal/tui/app.go | 260 |
| 22 | `(*App).projectInvite` | internal/cli/invitation_impl.go | 540 |

### Analysis by Component

#### CLI Layer (internal/cli/)
- 9 functions with high complexity
- Primary causes: Command-line flag parsing, input validation, user interaction
- **Recommendation**: Extract validation functions, use command pattern

#### TUI Layer (internal/tui/)
- 7 functions with high complexity
- Primary causes: Event handling, state updates, view rendering
- **Recommendation**: Split Update() methods into smaller handlers

#### Daemon API Layer (pkg/daemon/)
- 5 functions with high complexity
- Primary causes: Request validation, error handling, response formatting
- **Recommendation**: Extract validation and response builders

#### AWS Layer (pkg/aws/)
- 5 functions with high complexity
- Primary causes: AWS SDK interactions, error handling, retry logic
- **Recommendation**: Use strategy pattern for instance building

#### Core Business Logic (pkg/*)
- 14 functions with high complexity
- Spread across invitation, retry, state, types, etc.
- **Recommendation**: Component-specific refactoring strategies

### Complexity Categorization

**Command Handlers** (15 functions):
- Large switch/case statements for subcommands
- Multiple flag validations
- User confirmation prompts
- Recommended: Command pattern refactoring

**State Management** (12 functions):
- BubbleTea Update() methods
- Event multiplexing
- Asynchronous operation handling
- Recommended: Extract message handlers

**Request Handlers** (8 functions):
- REST API endpoint handlers
- JSON validation
- Database operations
- Recommended: Extract validation middleware

**Business Logic** (5 functions):
- Complex algorithms (invitation parsing, budget calculation)
- Multi-step workflows
- Recommended: Break into smaller composed functions

## TypeScript/React ESLint Issues

### Total: 97 problems (37 errors, 60 warnings)

#### Severity Distribution
- **Errors**: 37 (blocking for strict linting)
- **Warnings**: 60 (non-blocking but should be addressed)

### Issues by Category

#### 1. TypeScript Type Safety (24 errors)
**Issue**: `Unexpected any. Specify a different type` (@typescript-eslint/no-explicit-any)

**Locations**: src/App.tsx (22 instances), src/WebView.tsx (2 instances)

**Examples**:
```typescript
// Line 619
const handleSomething = (data: any) => { ... }  // ❌

// Line 3631
const processResult = (result: any) => { ... }  // ❌
```

**Impact**: Loss of type safety, potential runtime errors

**Recommendation**: Define proper TypeScript interfaces
```typescript
interface TemplateData {
  name: string;
  slug: string;
  description: string;
}

const handleTemplate = (data: TemplateData) => { ... }  // ✅
```

#### 2. Console Statements (47 warnings)
**Issue**: `Unexpected console statement` (no-console)

**Locations**: Widespread across src/App.tsx, src/Terminal.tsx, src/test-setup.ts

**Examples**:
```typescript
console.log('Debug info:', data);  // ❌
console.error('Error:', err);      // ❌
```

**Impact**: Production builds should not have console statements

**Recommendation**: Use proper logging library
```typescript
import { logger } from './utils/logger';
logger.debug('Debug info:', data);  // ✅
logger.error('Error:', err);        // ✅
```

#### 3. Unused Variables (6 errors)
**Issue**: Variables defined but never used (@typescript-eslint/no-unused-vars)

**Locations**:
- Line 1985: `'index' is defined but never used`
- Line 3826: `'result' is assigned a value but never used`
- Line 4557: `'projectBudget' is assigned a value but never used`
- Line 4889, 4898: `'item' is defined but never used`
- Line 5309: `'BudgetManagementView' is assigned a value but never used`
- Line 5712: `'buildModalVisible' is assigned a value but never used`

**Recommendation**: Remove unused code or prefix with underscore if intentional

#### 4. React Hooks Issues (5 warnings + 1 error)
**Issue**: Missing dependencies in useEffect (@typescript-eslint/react-hooks/exhaustive-deps)

**Locations**:
- Line 1473: Missing `loadApplicationData`
- Line 1560: Missing `loadApplicationData`
- Line 6029: Unnecessary dependency `state.marketplaceTemplates`
- Line 6700: Missing `fetchLogs`
- Line 4554: **ERROR** - React Hook called conditionally (react-hooks/rules-of-hooks)

**Examples**:
```typescript
// ❌ Missing dependency
useEffect(() => {
  loadApplicationData();
}, []);

// ✅ Correct dependencies
useEffect(() => {
  loadApplicationData();
}, [loadApplicationData]);

// ❌ Conditional hook call (CRITICAL ERROR)
if (condition) {
  const [state, setState] = React.useState();  // Hooks must be at top level
}
```

**Impact**: Potential stale closures, infinite loops, or React errors

#### 5. Code Complexity (3 warnings)
**Issue**: Functions exceed maximum lines/depth

**Locations**:
- Line 11 (Terminal.tsx): Arrow function has 160 lines (max 100)
- Line 19 (Terminal.tsx): Arrow function has 103 lines (max 100)
- Line 10 (WebView.tsx): Arrow function has 170 lines (max 100)
- Line 735 (App.tsx): Blocks nested too deeply (5 levels, max 4)

**Recommendation**: Extract sub-components and utility functions

#### 6. Code Quality Issues (3 errors)
**Issues**:
- Line 2039: Lexical declaration in case block (no-case-declarations)
- Line 13 (WebView.tsx): Calling impure function during render (`Date.now()`)

**Examples**:
```typescript
// ❌ Impure function in render
const [lastRefresh, setLastRefresh] = useState(Date.now());

// ✅ Use useRef or initialize outside render
const initialTime = useRef(Date.now());
```

### ESLint Issues by File

#### src/App.tsx (87 issues - 33 errors, 54 warnings)
Primary file containing most of the application logic. Large monolithic component.

**Breakdown**:
- 22 `@typescript-eslint/no-explicit-any` errors
- 47 `no-console` warnings
- 4 `@typescript-eslint/no-unused-vars` errors
- 3 React Hooks warnings
- 1 `max-depth` error
- 1 `no-case-declarations` error
- 1 `react-hooks/rules-of-hooks` **CRITICAL** error

#### src/Terminal.tsx (3 warnings)
- 2 `max-lines-per-function` warnings
- 1 `no-console` warning

#### src/WebView.tsx (3 warnings + 1 error)
- 1 `max-lines-per-function` warning
- 1 `react-hooks/purity` **ERROR** (impure function call)

#### src/test-setup.ts (2 warnings)
- 2 `no-console` warnings

## Prioritization for Future Work

### High Priority (Blocking for A+ rating)

#### Go
1. **Refactor CLI command handlers** (9 functions, complexity 20-33)
   - Extract flag parsing logic
   - Create command pattern structure
   - Split validation into separate functions

2. **Refactor TUI Update methods** (7 functions, complexity 19-29)
   - Extract message handlers
   - Create handler registry pattern
   - Split state updates into smaller functions

#### TypeScript
1. **Fix critical React Hooks error** (Line 4554)
   - Move hook call to component top level
   - Critical for React compliance

2. **Fix impure function call** (WebView.tsx:13)
   - Use useRef for initial timestamp
   - Prevents render instability

3. **Remove `any` types** (24 instances)
   - Define proper TypeScript interfaces
   - Restore type safety

### Medium Priority (Code quality improvement)

#### Go
1. Refactor daemon request handlers (5 functions, complexity 19-26)
2. Simplify AWS instance builder (1 function, complexity 28)

#### TypeScript
1. Remove unused variables (6 instances)
2. Fix React Hooks dependencies (4 instances)
3. Remove case block declarations (1 instance)

### Low Priority (Nice to have)

#### Go
1. Refactor remaining business logic functions (14 functions, complexity 16-19)

#### TypeScript
1. Replace console statements with logger (47 instances)
2. Split large components (3 functions > 100 lines)
3. Reduce nesting depth (1 instance, depth 5)

## Measurement

### Current State (v0.5.16)

**Go Report Card**: A+ (with 40 ignored cyclomatic complexity warnings)
**ESLint**: 97 problems (37 errors, 60 warnings)

### Target State (v0.6.0+)

**Go Report Card**: A+ (with < 20 ignored cyclomatic complexity warnings)
**ESLint**: < 20 problems (0 critical errors, < 20 warnings)

## Refactoring Strategy

### Phase 1: Critical Fixes (v0.5.17)
- Fix React Hooks critical error (WebView/App)
- Fix impure function call (WebView)
- Remove top 5 unused variables

### Phase 2: Type Safety (v0.6.0)
- Define core TypeScript interfaces
- Remove all `any` types from API calls
- Remove all `any` types from event handlers

### Phase 3: Code Structure (v0.6.1)
- Refactor CLI command handlers (command pattern)
- Refactor TUI Update methods (handler registry)
- Split large React components

### Phase 4: Polish (v0.6.2+)
- Implement proper logging system
- Remove console statements
- Fix React Hooks dependencies
- Reduce Go complexity in remaining functions

## Notes

- **Not Blocking v0.5.16**: All tests pass, functionality works correctly
- **Technical Debt**: Issues documented for future cleanup
- **No New Linting Issues**: Enforce linting in CI for new code
- **Gradual Improvement**: Address during feature development, not as dedicated effort

## References

- Cyclomatic Complexity Report: `/tmp/gocyclo_results.txt`
- ESLint Report: `/tmp/eslint_results.txt`
- Go Report Card: https://goreportcard.com/report/github.com/scttfrdmn/prism
