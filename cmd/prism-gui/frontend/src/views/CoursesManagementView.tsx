import { CoursesPanel } from '../components/CoursesPanel'

// Top-level component (not inside PrismApp) to prevent re-mount on state change (#13).
export function CoursesManagementView() {
  return <CoursesPanel />
}
