import mineflayer from 'mineflayer'
import pathfinderPkg from 'mineflayer-pathfinder'

const { pathfinder, Movements, goals } = pathfinderPkg

export async function connectBot({ host, port, username, version, auth, log }) {
  const options = { host, port, username, auth, hideErrors: false }
  if (version) {
    options.version = version
  }
  const bot = mineflayer.createBot(options)
  bot.loadPlugin(pathfinder)

  await new Promise((resolve, reject) => {
    const onSpawn = () => {
      cleanup()
      resolve()
    }
    const onError = (err) => {
      cleanup()
      reject(err)
    }
    const onEnd = (reason) => {
      cleanup()
      reject(new Error(`bot disconnected before spawn: ${reason}`))
    }
    function cleanup() {
      bot.off('spawn', onSpawn)
      bot.off('error', onError)
      bot.off('end', onEnd)
    }
    bot.once('spawn', onSpawn)
    bot.once('error', onError)
    bot.once('end', onEnd)
  })

  const mcData = (await import('minecraft-data')).default(bot.version)
  const movements = new Movements(bot, mcData)
  bot.pathfinder.setMovements(movements)

  log('bot connected', { username, auth, bot_version: bot.version, target_version: version || 'auto' })
  return bot
}

export function snapshotPosition(bot) {
  return {
    x: Number(bot.entity.position.x.toFixed(3)),
    y: Number(bot.entity.position.y.toFixed(3)),
    z: Number(bot.entity.position.z.toFixed(3)),
  }
}

export async function gotoPoint(bot, target, range = 1.0, timeoutMs = 30000) {
  const goal = new goals.GoalNear(target.x, target.y, target.z, range)
  bot.pathfinder.setGoal(goal)

  await waitFor(
    () => bot.entity.position.distanceTo({ x: target.x, y: target.y, z: target.z }) <= range,
    timeoutMs,
    100,
    `failed to reach target ${target.x},${target.y},${target.z}`,
  )

  bot.pathfinder.setGoal(null)
}

export async function waitFor(predicate, timeoutMs, intervalMs, timeoutMessage) {
  const started = Date.now()
  while (Date.now() - started < timeoutMs) {
    if (predicate()) {
      return
    }
    await sleep(intervalMs)
  }
  throw new Error(timeoutMessage)
}

export async function sleep(ms) {
  await new Promise((resolve) => setTimeout(resolve, ms))
}
