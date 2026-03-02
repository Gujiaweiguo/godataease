export const validateOrigin = (origin: string): { isValid: boolean; error?: string } => {
  try {
    const url = new URL(origin)
    if (!['http:', 'https:'].includes(url.protocol)) {
      return { isValid: false, error: 'Invalid protocol' }
    }
    return { isValid: true }
  } catch (_error) {
    return { isValid: false, error: 'Invalid origin URL' }
  }
}

const wildcardToRegex = (pattern: string): RegExp => {
  const escaped = pattern.replace(/[.+?^${}()|[\]\\]/g, '\\$&').replace(/\*/g, '.*')
  return new RegExp(`^${escaped}$`)
}

export const isOriginAllowed = (
  origin: string,
  allowedOrigins: string[] = [],
  allowWhenTokenMissing = false
): boolean => {
  if (allowWhenTokenMissing) {
    return true
  }
  if (!allowedOrigins?.length) {
    return false
  }

  return allowedOrigins.some(allowed => {
    if (allowed.includes('*')) {
      return wildcardToRegex(allowed).test(origin)
    }
    return allowed === origin
  })
}
