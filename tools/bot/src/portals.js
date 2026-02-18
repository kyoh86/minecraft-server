import fs from 'node:fs'
import yaml from 'js-yaml'

function parseTriple(value) {
  const [x, y, z] = value.split(',').map((v) => Number(v))
  return { x, y, z }
}

function parseLocation(location) {
  const [world, from, to] = location.split(':')
  const p1 = parseTriple(from)
  const p2 = parseTriple(to)
  const min = {
    x: Math.min(p1.x, p2.x),
    y: Math.min(p1.y, p2.y),
    z: Math.min(p1.z, p2.z),
  }
  const max = {
    x: Math.max(p1.x, p2.x),
    y: Math.max(p1.y, p2.y),
    z: Math.max(p1.z, p2.z),
  }
  const center = {
    x: (min.x + max.x) / 2,
    y: (min.y + max.y) / 2,
    z: (min.z + max.z) / 2,
  }
  return { world, min, max, center }
}

export function loadPortals(portalsPath) {
  const raw = fs.readFileSync(portalsPath, 'utf8')
  const parsed = yaml.load(raw)
  const portals = parsed?.portals || {}
  const result = {}
  for (const [name, value] of Object.entries(portals)) {
    result[name] = {
      ...value,
      bounds: parseLocation(value.location),
    }
  }
  return result
}

export function inBounds(position, bounds) {
  return (
    position.x >= bounds.min.x &&
    position.x < bounds.max.x + 1 &&
    position.y >= bounds.min.y &&
    position.y < bounds.max.y + 1 &&
    position.z >= bounds.min.z &&
    position.z < bounds.max.z + 1
  )
}
