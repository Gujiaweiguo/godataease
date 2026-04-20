const expWarningMs = 10000

type BootstrapSessionState = {
  exp?: number | null
  isDesktop?: boolean | null
  now?: number
  time?: number | null
  token?: string | null
}

export const isBootstrapSessionValid = ({
  token,
  exp,
  isDesktop,
  now = Date.now()
}: BootstrapSessionState) => {
  if (isDesktop) {
    return true
  }
  if (!token || !exp) {
    return false
  }
  if (exp - now < expWarningMs) {
    return false
  }
  return true
}
