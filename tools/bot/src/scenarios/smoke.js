import { snapshotPosition } from '../bot.js'
import { runPortalResourceToMainhall } from './portal_resource_to_mainhall.js'
import { runWorldTransferCommandRoundtrip } from './world_transfer_command_roundtrip.js'

export async function runSmoke({ bot, portals, log, report, addCheck }) {
  log('smoke: start world transfer roundtrip')
  const transfer = await runWorldTransferCommandRoundtrip({ bot, portals, log, report, addCheck })

  log('smoke: start portal roundtrip')
  const portal = await runPortalResourceToMainhall({ bot, portals, log, report, addCheck })

  addCheck(report, 'smoke_completed', true, {
    steps: ['world_transfer_command_roundtrip', 'portal_resource_to_mainhall'],
  })

  return {
    start: transfer?.start || snapshotPosition(bot),
    end: portal?.end || snapshotPosition(bot),
  }
}
