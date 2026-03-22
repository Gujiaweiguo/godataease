const expWarningMs = 10000
const sessionAgeWarningMs = 90000

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
  time,
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
  if (!time) {
    return true
  }
  return now - time <= sessionAgeWarningMs
}
