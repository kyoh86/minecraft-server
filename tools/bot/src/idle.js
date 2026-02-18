import { connectBot, snapshotPosition } from './bot.js'
import { loadConfig } from './config.js'
import { createLogger } from './logger.js'

const cfg = loadConfig()
const logger = createLogger()

async function main() {
  const bot = await connectBot({
    host: cfg.host,
    port: cfg.port,
    username: cfg.username,
    version: cfg.version,
    auth: cfg.auth,
    log: logger.log,
  })

  setInterval(() => {
    logger.log('heartbeat', { pos: snapshotPosition(bot) })
  }, 10000)

  bot.on('end', (reason) => {
    logger.log('bot ended', { reason })
    process.exit(0)
  })

  bot.on('kicked', (reason) => {
    logger.log('bot kicked', { reason: String(reason) })
  })
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
