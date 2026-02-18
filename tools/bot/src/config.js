import path from 'node:path'

export function loadConfig() {
  const host = process.env.BOT_HOST || 'localhost'
  const port = Number(process.env.BOT_PORT || 25565)
  const username = process.env.BOT_USERNAME || 'codexbot'
  const version = process.env.BOT_VERSION || ''
  const auth = process.env.BOT_AUTH || 'offline'
  const scenario = process.env.BOT_SCENARIO || ''
  const reportDir = process.env.BOT_REPORT_DIR || path.resolve(process.cwd(), '../../runtime/bot-reports')
  const portalsPath = process.env.BOT_PORTALS_PATH || '/runtime/world/plugins/Multiverse-Portals/portals.yml'

  return { host, port, username, version, auth, scenario, reportDir, portalsPath }
}
