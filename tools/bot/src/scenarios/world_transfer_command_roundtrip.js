import { sleep, snapshotPosition, waitFor } from '../bot.js'

export async function runWorldTransferCommandRoundtrip({ bot, portals, log, report, addCheck }) {
  const toResource = portals.gate_resource

  log('command transfer to resource', { command: '/mvtp resource' })
  bot.chat('/mvtp resource')
  await sleep(1200)
  await waitFor(
    () => Math.abs(bot.entity.position.y - 106) < 8,
    15000,
    250,
    'timeout waiting move to resource by command',
  )
  const resourcePos = snapshotPosition(bot)
  addCheck(report, 'mvtp_to_resource', Math.abs(resourcePos.y - 106) < 8, {
    expected_y: 106,
    actual: resourcePos,
  })

  log('command transfer to mainhall', { command: '/mvtp mainhall' })
  bot.chat('/mvtp mainhall')
  await sleep(1200)
  await waitFor(
    () => Math.abs(bot.entity.position.y - toResource.bounds.center.y) < 8,
    15000,
    250,
    'timeout waiting move to mainhall by command',
  )
  const mainhallPos = snapshotPosition(bot)
  addCheck(report, 'mvtp_to_mainhall', Math.abs(mainhallPos.y - toResource.bounds.center.y) < 8, {
    expected_y: toResource.bounds.center.y,
    actual: mainhallPos,
  })

  return { start: resourcePos, end: mainhallPos }
}
