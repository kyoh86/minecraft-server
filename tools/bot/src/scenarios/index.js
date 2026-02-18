import { runPortalResourceToMainhall } from './portal_resource_to_mainhall.js'
import { runSmoke } from './smoke.js'
import { runWorldTransferCommandRoundtrip } from './world_transfer_command_roundtrip.js'

export const scenarios = {
  portal_resource_to_mainhall: runPortalResourceToMainhall,
  smoke: runSmoke,
  world_transfer_command_roundtrip: runWorldTransferCommandRoundtrip,
}
