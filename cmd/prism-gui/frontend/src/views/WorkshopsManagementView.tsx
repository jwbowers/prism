import { WorkshopsPanel } from '../components/WorkshopsPanel'

// Top-level component (not inside PrismApp) to prevent re-mount on state change (#13).
export function WorkshopsManagementView() {
  return <WorkshopsPanel />
}
