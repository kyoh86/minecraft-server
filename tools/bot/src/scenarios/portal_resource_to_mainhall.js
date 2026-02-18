import { gotoPoint, sleep, snapshotPosition, waitFor } from '../bot.js'

export async function runPortalResourceToMainhall({ bot, portals, log, report, addCheck }) {
  const toResource = portals.gate_resource
  const toMainhall = portals.gate_resource_to_mainhall
  const portalDebug = process.env.BOT_PORTAL_DEBUG === '1'

  if (!toResource || !toMainhall) {
    throw new Error('required portals are missing: gate_resource or gate_resource_to_mainhall')
  }

  log('setup world to resource by command', { command: '/mvtp resource' })
  bot.chat('/mvtp resource')
  await sleep(1200)

  await waitFor(
    () => Math.abs(bot.entity.position.y - toMainhall.bounds.center.y) < 8,
    30000,
    250,
    'timeout waiting teleport to resource',
  )
  const inResourcePos = snapshotPosition(bot)
  addCheck(report, 'setup_resource_world', Math.abs(inResourcePos.y - toMainhall.bounds.center.y) < 8, {
    expected_y: toMainhall.bounds.center.y,
    actual: inResourcePos,
  })
  bot.chat('/mvp debug off')
  await sleep(300)
  if (portalDebug) {
    bot.chat('/mvp debug on')
    await sleep(300)
  }

  // Move toward the portal plane without crossing the wall behind it (z=-9).
  log('walk to resource portal front', { target: { x: 0.5, y: 107, z: -7.2 } })
  await gotoPoint(bot, { x: 0.5, y: 107, z: -7.2 }, 0.8, 45000)
  const stagedPos = snapshotPosition(bot)
  addCheck(report, 'stage_near_resource_portal', stagedPos.z <= -6 && stagedPos.y >= 106, {
    expected: { min_y: 106, max_z: -6 },
    actual: stagedPos,
  })

  // Enter portal area with movement packets (not /tp), then keep slight motion.
  log('walk into portal plane', { target: { x: 0.5, y: 107, z: -7.95 } })
  let teleportedEarly = false
  try {
    await gotoPoint(bot, { x: 0.5, y: 107, z: -7.95 }, 0.5, 45000)
    await sleep(6000)
  } catch (err) {
    const posAfterError = snapshotPosition(bot)
    if (Math.abs(posAfterError.y - toResource.bounds.center.y) < 8) {
      teleportedEarly = true
      log('teleported while approaching portal', { pos: posAfterError })
    } else {
      throw err
    }
  }

  if (!teleportedEarly) {
    await waitFor(
      () => Math.abs(bot.entity.position.y - toResource.bounds.center.y) < 8,
      30000,
      250,
      'timeout waiting teleport to mainhall',
    )
  }
  const endPos = snapshotPosition(bot)

  const arrivedMainhall = Math.abs(endPos.y - toResource.bounds.center.y) < 8
  addCheck(report, 'portal_resource_to_mainhall', arrivedMainhall, {
    expected_y: toResource.bounds.center.y,
    actual: endPos,
  })

  const nearHub = Math.abs(endPos.x) <= 16 && Math.abs(endPos.z) <= 16
  addCheck(report, 'arrival_near_mainhall_hub', nearHub, {
    expected: { xz_abs_max: 16 },
    actual: endPos,
  })

  return { start: inResourcePos, end: endPos }
}
