import { connectBot, snapshotPosition } from './bot.js'
import { loadConfig } from './config.js'
import { createLogger } from './logger.js'
import { loadPortals } from './portals.js'
import { addCheck, createReport, finalizeReport, writeReport } from './report.js'
import { scenarios } from './scenarios/index.js'

const cfg = loadConfig()
const logger = createLogger()
const scenarioName = cfg.scenario

if (!scenarioName) {
  console.error('BOT_SCENARIO is required')
  process.exit(1)
}

const scenario = scenarios[scenarioName]
if (!scenario) {
  console.error(`unknown scenario: ${scenarioName}`)
  process.exit(1)
}

const username = process.env.BOT_USERNAME || 'codexbot'
let bot = null
const report = createReport({ scenario: scenarioName, logger })

async function main() {
  const portals = loadPortals(cfg.portalsPath)
  bot = await connectBot({
    host: cfg.host,
    port: cfg.port,
    username,
    version: cfg.version,
    auth: cfg.auth,
    log: logger.log,
  })
  bot.on('messagestr', (msg) => {
    logger.log('server message', { msg })
  })
  bot.on('kicked', (reason) => {
    logger.log('kicked', { reason: String(reason) })
  })

  const initial = snapshotPosition(bot)
  report.start = {
    dimension: 'overworld',
    pos: initial,
  }

  const result = await scenario({
    bot,
    portals,
    log: logger.log,
    report,
    addCheck,
  })

  const passed = report.checks.every((c) => c.pass)
  finalizeReport(report, {
    end: { dimension: 'overworld', pos: result.end },
    result: passed ? 'pass' : 'fail',
    error: null,
  })

  const file = writeReport(cfg.reportDir, report)
  logger.log('scenario finished', { file, result: report.result, duration_ms: report.duration_ms })
  console.log(`report: ${file}`)
  if (!passed) {
    process.exitCode = 1
  }
}

main()
  .catch((err) => {
    logger.log('scenario error', { error: err.message })
    const endPos = bot ? snapshotPosition(bot) : null
    finalizeReport(report, {
      end: { dimension: 'overworld', pos: endPos },
      result: 'fail',
      error: err.message,
    })
    const file = writeReport(cfg.reportDir, report)
    console.error(`report: ${file}`)
    console.error(err)
    process.exitCode = 1
  })
  .finally(() => {
    if (bot) {
      bot.quit('test done')
    }
  })
