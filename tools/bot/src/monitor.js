import { connectBot } from './bot.js'
import { loadConfig } from './config.js'
import { createLogger } from './logger.js'
import { runScenarioOnce } from './scenario_runner.js'

const cfg = loadConfig()
const logger = createLogger()
const scenarioName = process.env.BOT_MONITOR_SCENARIO || 'smoke'
const intervalSec = Number(process.env.BOT_MONITOR_INTERVAL_SEC || 300)

if (!Number.isFinite(intervalSec) || intervalSec <= 0) {
  console.error('BOT_MONITOR_INTERVAL_SEC must be a positive number')
  process.exit(1)
}

let stopping = false
let bot = null

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function runLoop() {
  while (!stopping) {
    const started = Date.now()
    const { report, file, error } = await runScenarioOnce({
      name: scenarioName,
      bot,
      logger,
      reportDir: cfg.reportDir,
      portalsPath: cfg.portalsPath,
    })
    logger.log('monitor scenario finished', {
      scenario: scenarioName,
      result: report.result,
      report: file,
      duration_ms: report.duration_ms,
      error: error ? error.message : null,
    })
    const elapsed = Date.now() - started
    const waitMs = Math.max(1000, intervalSec * 1000 - elapsed)
    await sleep(waitMs)
  }
}

async function main() {
  bot = await connectBot({
    host: cfg.host,
    port: cfg.port,
    username: cfg.username,
    version: cfg.version,
    auth: cfg.auth,
    log: logger.log,
  })
  logger.log('monitor started', { scenario: scenarioName, interval_sec: intervalSec })
  await runLoop()
}

process.on('SIGINT', () => {
  stopping = true
})
process.on('SIGTERM', () => {
  stopping = true
})

main()
  .catch((err) => {
    console.error(err)
    process.exitCode = 1
  })
  .finally(() => {
    if (bot) {
      bot.quit('monitor stop')
      bot = null
    }
  })
