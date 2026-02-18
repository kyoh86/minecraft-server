import { connectBot } from './bot.js'
import { loadConfig } from './config.js'
import { createLogger } from './logger.js'
import { scenarios } from './scenarios/index.js'
import { runScenarioOnce } from './scenario_runner.js'

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

async function main() {
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

  const { report, file, passed, error } = await runScenarioOnce({
    name: scenarioName,
    bot,
    logger,
    reportDir: cfg.reportDir,
    portalsPath: cfg.portalsPath,
  })
  logger.log('scenario finished', { file, result: report.result, duration_ms: report.duration_ms })
  console.log(`report: ${file}`)
  if (error) {
    throw error
  }
  if (!passed) {
    process.exitCode = 1
  }
}

main()
  .catch((err) => {
    console.error(err)
    process.exitCode = 1
  })
  .finally(() => {
    if (bot) {
      bot.quit('test done')
    }
  })
