export const resolveEmbeddedOrigin = (): string => {
  const referrer = document.referrer
  if (window.self !== window.top && referrer) {
    try {
      return new URL(referrer).origin
    } catch (error) {
      return ''
    }
  }
  return window.location.origin
}

const normalizeOrigin = (origin: string): string => {
  if (!origin) {
    return ''
  }
  let normalized = origin.trim()
  while (normalized.endsWith('/')) {
    normalized = normalized.slice(0, -1)
  }
  return normalized
}

const getOriginHost = (origin: string): string => {
  try {
    return new URL(origin).hostname
  } catch (error) {
    return ''
  }
}

export const isAllowedEmbeddedMessageOrigin = (
  origin: string,
  allowlist: string[] = [],
  enforceAllowlist = false
): boolean => {
  const normalizedOrigin = normalizeOrigin(origin)
  if (!normalizedOrigin) {
    return false
  }
  if (enforceAllowlist) {
    if (!allowlist.length) {
      return false
    }
    const originHost = getOriginHost(normalizedOrigin)
    return allowlist.some(entry => {
      const normalizedEntry = normalizeOrigin(entry)
      if (!normalizedEntry) {
        return false
      }
      if (normalizedEntry.includes('://')) {
        return normalizedEntry === normalizedOrigin
      }
      return Boolean(originHost) && normalizedEntry === originHost
    })
  }
  const expectedOrigin = normalizeOrigin(resolveEmbeddedOrigin())
  if (!expectedOrigin) {
    return true
  }
  return normalizedOrigin === expectedOrigin
}
