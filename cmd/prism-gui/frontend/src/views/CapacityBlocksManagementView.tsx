import CapacityBlocksPanel from '../components/CapacityBlocksPanel'

// Top-level component for EC2 Capacity Blocks (v0.20.0 #63).
// Top-level (not inside PrismApp) to prevent re-mount on state change (#13).
export function CapacityBlocksManagementView() {
  return <CapacityBlocksPanel />
}
