import { describe, expect, it, vi } from 'vitest'

vi.mock('@/hooks/web/useI18n', () => ({
  useI18n: () => ({ t: (value: string) => value })
}))

vi.mock('@/store/modules/app', () => ({
  useAppStoreWithOut: () => ({ getIsDataEaseBi: false })
}))

vi.mock('@/store/modules/map', () => ({
  useMapStoreWithOut: () => ({})
}))

vi.mock('@/store/modules/link', () => ({
  useLinkStoreWithOut: () => ({})
}))

vi.mock('@/config/axios', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn()
  }
}))

describe('chart util import cycle', () => {
  it('imports chart defaults without formatter initialization errors', async () => {
    const mod = await import('@/views/chart/components/editor/util/chart')

    expect(mod.DEFAULT_LABEL.labelFormatter).toBeDefined()
    expect(mod.DEFAULT_TOOLTIP.tooltipFormatter).toBeDefined()
    expect(mod.DEFAULT_TITLE_STYLE.remarkBackgroundColor).toBe('#ffffff')
  })
})
