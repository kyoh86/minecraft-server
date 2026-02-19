import { runPortalResourceToMainhall } from './portal_resource_to_mainhall.js'
import { runWorldTransferCommandRoundtrip } from './world_transfer_command_roundtrip.js'

export const scenarios = {
  portal_resource_to_mainhall: runPortalResourceToMainhall,
  world_transfer_command_roundtrip: runWorldTransferCommandRoundtrip,
}
