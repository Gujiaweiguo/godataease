export const shouldInitializeDashboardCreate = (resourceId?: unknown, opt?: unknown) => {
  return opt === 'create' || (!resourceId && !opt)
}

export const createAsyncLoadGate = () => {
  let loaded = false
  let resolveLoaded: (() => void) | null = null

  const wait = new Promise<void>(resolve => {
    resolveLoaded = resolve
  })

  const markLoaded = () => {
    if (loaded) {
      return
    }
    loaded = true
    resolveLoaded?.()
    resolveLoaded = null
  }

  return {
    wait,
    markLoaded,
    isLoaded: () => loaded
  }
}
