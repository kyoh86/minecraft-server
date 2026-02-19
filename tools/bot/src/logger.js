export function createLogger() {
  const logs = []

  function log(message, extra = {}) {
    const entry = {
      at: new Date().toISOString(),
      message,
      ...extra,
    }
    logs.push(entry)
    console.log(`[bot] ${entry.at} ${message}`, Object.keys(extra).length ? extra : '')
  }

  return { logs, log }
}
