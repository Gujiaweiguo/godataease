import { useEmitt } from '@/hooks/web/useEmitt'
import { performLogout } from '@/utils/logout'

export type MenuActionHandler = (config?: Record<string, unknown>) => void

const menuActionHandlers: Record<string, MenuActionHandler> = {
  'open-about-dialog': () => {
    useEmitt().emitter.emit('open-about-dialog')
  },
  'open-language-selector': () => {
  },
  'user-logout': () => {
    performLogout()
  },
  'data-export-center': () => {
    useEmitt().emitter.emit('data-export-center')
  }
}

export function executeMenuAction(event: string, config?: Record<string, unknown>): void {
  const handler = menuActionHandlers[event]
  if (handler) {
    handler(config)
  } else {
    console.warn(`Unknown menu action: ${event}`)
  }
}

export function isEventMenu(menuType?: string): boolean {
  return menuType === 'event'
}

export function isExternalMenu(menuType?: string): boolean {
  return menuType === 'external'
}

export function isRouteMenu(menuType?: string): boolean {
  return !menuType || menuType === 'route'
}
