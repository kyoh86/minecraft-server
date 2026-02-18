import readline from 'node:readline'
import { connectBot, snapshotPosition } from './bot.js'
import { loadConfig } from './config.js'
import { createLogger } from './logger.js'
import { runScenarioOnce } from './scenario_runner.js'
import { scenarios } from './scenarios/index.js'

const cfg = loadConfig()
const logger = createLogger({ stderr: true })

let bot = null

function emit(obj) {
  process.stdout.write(`${JSON.stringify(obj)}\n`)
}

function ok(id, data = {}) {
  emit({ id, ok: true, ...data })
}

function fail(id, code, message, data = {}) {
  emit({ id, ok: false, error: { code, message }, ...data })
}

function ensureConnected() {
  if (!bot) {
    throw new Error('bot is not connected')
  }
}

async function actionConnect(id, req) {
  if (bot) {
    ok(id, {
      connected: true,
      username: bot.username,
      pos: snapshotPosition(bot),
      bot_version: bot.version,
    })
    return
  }

  const username = req.username || cfg.username
  const host = req.host || cfg.host
  const port = Number(req.port || cfg.port)
  const version = req.version || cfg.version
  const auth = req.auth || cfg.auth
  bot = await connectBot({
    host,
    port,
    username,
    version,
    auth,
    log: logger.log,
  })
  bot.on('kicked', (reason) => {
    logger.log('kicked', { reason: String(reason) })
  })
  bot.on('end', (reason) => {
    logger.log('bot ended', { reason: String(reason) })
    bot = null
  })

  ok(id, {
    connected: true,
    username: bot.username,
    pos: snapshotPosition(bot),
    bot_version: bot.version,
  })
}

async function actionDisconnect(id) {
  if (!bot) {
    ok(id, { connected: false })
    return
  }
  bot.quit('control disconnect')
  bot = null
  ok(id, { connected: false })
}

async function actionStatus(id) {
  if (!bot) {
    ok(id, { connected: false })
    return
  }
  ok(id, {
    connected: true,
    username: bot.username,
    pos: snapshotPosition(bot),
    bot_version: bot.version,
  })
}

async function actionSnapshot(id) {
  ensureConnected()
  ok(id, {
    connected: true,
    pos: snapshotPosition(bot),
  })
}

async function actionRunScenario(id, req) {
  ensureConnected()
  const name = String(req.name || '').trim()
  if (!name) {
    fail(id, 'invalid_request', 'name is required')
    return
  }
  const scenario = scenarios[name]
  if (!scenario) {
    fail(id, 'invalid_request', `unknown scenario: ${name}`)
    return
  }

  const { report, file, error } = await runScenarioOnce({
    name,
    bot,
    logger,
    reportDir: cfg.reportDir,
    portalsPath: cfg.portalsPath,
  })
  if (error) {
    fail(id, 'scenario_failed', error.message, {
      scenario: name,
      report: file,
      checks: report.checks,
    })
    return
  }
  ok(id, {
    scenario: name,
    result: report.result,
    report: file,
    checks: report.checks,
  })
}

async function actionQuit(id) {
  if (bot) {
    bot.quit('control quit')
    bot = null
  }
  ok(id, { quitting: true })
  process.exit(0)
}

async function handleLine(line) {
  const trimmed = line.trim()
  if (!trimmed) {
    return
  }

  let req
  try {
    req = JSON.parse(trimmed)
  } catch (err) {
    fail('', 'invalid_json', err.message)
    return
  }
  const id = String(req.id || '')
  const action = String(req.action || '')

  try {
    switch (action) {
      case 'connect':
        await actionConnect(id, req)
        return
      case 'disconnect':
        await actionDisconnect(id)
        return
      case 'status':
        await actionStatus(id)
        return
      case 'snapshot':
        await actionSnapshot(id)
        return
      case 'runScenario':
        await actionRunScenario(id, req)
        return
      case 'quit':
        await actionQuit(id)
        return
      default:
        fail(id, 'invalid_request', `unknown action: ${action}`)
    }
  } catch (err) {
    fail(id, 'internal_error', err.message)
  }
}

function start() {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: false,
  })

  ok('ready', {
    actions: ['connect', 'disconnect', 'status', 'snapshot', 'runScenario', 'quit'],
  })

  let pending = Promise.resolve()
  rl.on('line', (line) => {
    pending = pending.then(() => handleLine(line)).catch((err) => {
      fail('', 'internal_error', err.message)
    })
  })

  rl.on('close', async () => {
    await pending
    if (bot) {
      bot.quit('stdin closed')
      bot = null
    }
    process.exit(0)
  })
}

start()
