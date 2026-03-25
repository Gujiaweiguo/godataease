export interface InteractiveBootstrapStore {
  clear: () => void
  initInteractive: (refresh?: boolean) => Promise<void>
}

export const bootstrapInteractiveInBackground = (
  interactiveStore: InteractiveBootstrapStore,
  refresh = true
) => {
  interactiveStore.clear()
  void interactiveStore.initInteractive(refresh).catch(() => undefined)
}
