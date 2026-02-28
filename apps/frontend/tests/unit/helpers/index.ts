import { vi } from 'vitest'

type WsCacheOverrides = Partial<{
  get: ReturnType<typeof vi.fn>
  set: ReturnType<typeof vi.fn>
  delete: ReturnType<typeof vi.fn>
}>

type ApiMockFn = ReturnType<typeof vi.fn>

type UtilsModuleMock = {
  setTitle: ApiMockFn
  isBtnShow: (val: string) => boolean
  nameTrim: ApiMockFn
}

type ChartEditorUtilModuleMock = {
  BASE_VIEW_CONFIG: Record<string, unknown>
  DEFAULT_INDICATOR_NAME_STYLE: Record<string, unknown>
  DEFAULT_INDICATOR_STYLE: Record<string, unknown>
  SENIOR_STYLE_SETTING_LIGHT: Record<string, unknown>
  getViewConfig: ApiMockFn
}

type FormatterModuleMock = {
  formatterItem: Record<string, unknown>
}

type ComponentListModuleMock = {
  default: unknown[]
  ACTION_SELECTION: Record<string, unknown>
  BASE_CAROUSEL: Record<string, unknown>
  BASE_EVENTS: { checked: boolean; type: string }
  COMMON_COMPONENT_BACKGROUND_DARK: Record<string, unknown>
  COMMON_COMPONENT_BACKGROUND_LIGHT: Record<string, unknown>
  COMMON_TAB_TITLE_BACKGROUND: Record<string, unknown>
  MULTI_DIMENSIONAL: Record<string, unknown>
  defaultStyleValue: Record<string, unknown>
  findBaseDeFaultAttr: ApiMockFn
}

type RouterChildMock = {
  path: string
  name: string
  hidden?: boolean
  children?: RouterChildMock[]
}

type RouterMock = {
  default: {
    beforeEach: ApiMockFn
    afterEach: ApiMockFn
    addRoute: ApiMockFn
    push: ApiMockFn
  }
  routes: RouterChildMock[]
}

type EmittEmitterMock = {
  emit: ApiMockFn
  on: ApiMockFn
  off: ApiMockFn
}

type EventBusModuleMock = {
  default: {
    emit: ApiMockFn
  }
}

type ModelUtilModuleMock = {
  isDesktop: ApiMockFn
}

type DatasetFormUtilModuleMock = {
  guid: ApiMockFn
}

type DataVisualizationEditorUtilModuleMock = {
  DEFAULT_CANVAS_STYLE_DATA_DARK: Record<string, unknown>
  DEFAULT_CANVAS_STYLE_DATA_LIGHT: Record<string, unknown>
  DEFAULT_CANVAS_STYLE_DATA_SCREEN_DARK: Record<string, unknown>
}

type PanelModuleMock = {
  default: Record<string, unknown>
}

type ViewUtilsModuleMock = {
  viewFieldTimeTrans: ApiMockFn
}

type ComponentUtilsModuleMock = {
  filterEnumParams: ApiMockFn
  filterEnumParamsReduce: ApiMockFn
  filterParamsOptions: ApiMockFn
}

type EmbeddedApiModuleMock = {
  embeddedInitIframeApi: ApiMockFn
  embeddedGetTokenArgsApi: ApiMockFn
}

type DvMainStoreWithOutMock = {
  curOriginThemes: string
  canvasAttachInfo: Record<string, unknown>
  setComponentData: ApiMockFn
  setCanvasStyle: ApiMockFn
  updateCurDvInfo: ApiMockFn
  setCanvasViewInfo: ApiMockFn
  setNowPanelTrackInfo: ApiMockFn
  setNowPanelJumpInfo: ApiMockFn
  updateDvInfoCall: ApiMockFn
  removeGroupArea: ApiMockFn
}

type SnapshotStoreWithOutMock = {
  resetStyleChangeTimes: ApiMockFn
}

type AppearanceStoreWithOutMock = {
  setCurrentFont: ApiMockFn
}

type AppStoreWithOutMock = {
  getIsIframe: boolean
}

type PermissionModuleMock = {
  pathValid: ApiMockFn
}

export const createUseCacheModuleMock = (overrides?: WsCacheOverrides) => {
  return {
    useCache: () => ({
      wsCache: {
        get: vi.fn(),
        set: vi.fn(),
        delete: vi.fn(),
        ...overrides
      }
    })
  }
}

export const createApiModuleMock = <T extends string>(methods: T[]) => {
  return methods.reduce(
    (acc, method) => {
      acc[method] = vi.fn()
      return acc
    },
    {} as Record<T, ApiMockFn>
  )
}

export const createResolvedApiModuleMock = <T extends string>(responses: Record<T, unknown>) => {
  const result = {} as Record<T, ApiMockFn>
  ;(Object.keys(responses) as T[]).forEach(method => {
    result[method] = vi.fn().mockResolvedValue(responses[method])
  })
  return result
}

export const createUtilsModuleMock = (overrides?: Partial<UtilsModuleMock>) => {
  const moduleMock: UtilsModuleMock = {
    setTitle: vi.fn(),
    isBtnShow: (val: string) => val === '1' || val === 'true',
    nameTrim: vi.fn(),
    ...overrides
  }
  return moduleMock
}

export const createChartEditorUtilModuleMock = (
  overrides?: Partial<ChartEditorUtilModuleMock>
) => {
  const moduleMock: ChartEditorUtilModuleMock = {
    BASE_VIEW_CONFIG: {},
    DEFAULT_INDICATOR_NAME_STYLE: {},
    DEFAULT_INDICATOR_STYLE: {},
    SENIOR_STYLE_SETTING_LIGHT: {},
    getViewConfig: vi.fn(() => ({ title: 'Test View', render: 'antv' })),
    ...overrides
  }
  return moduleMock
}

export const createFormatterModuleMock = (overrides?: Partial<FormatterModuleMock>) => {
  return {
    formatterItem: {},
    ...overrides
  }
}

export const createComponentListModuleMock = (overrides?: Partial<ComponentListModuleMock>) => {
  return {
    default: [],
    ACTION_SELECTION: {},
    BASE_CAROUSEL: {},
    BASE_EVENTS: { checked: false, type: 'displayChange' },
    COMMON_COMPONENT_BACKGROUND_DARK: {},
    COMMON_COMPONENT_BACKGROUND_LIGHT: {},
    COMMON_TAB_TITLE_BACKGROUND: {},
    MULTI_DIMENSIONAL: {},
    defaultStyleValue: {},
    findBaseDeFaultAttr: vi.fn(),
    ...overrides
  }
}

export const createRouterModuleMock = (overrides?: Partial<RouterMock>) => {
  return {
    default: {
      beforeEach: vi.fn(),
      afterEach: vi.fn(),
      addRoute: vi.fn(),
      push: vi.fn()
    },
    routes: [
      { path: '/', name: 'Root', children: [] },
      { path: '/login', name: 'Login', children: [] }
    ],
    ...overrides
  }
}

export const createRouterEstablishModuleMock = () => {
  return {
    generateRoutesFn2: vi.fn((routers: RouterChildMock[]) => {
      return routers.map(router => ({
        path: router.path,
        name: router.name,
        hidden: router.hidden,
        children:
          router.children?.map(child => ({
            path: child.path,
            name: child.name,
            hidden: child.hidden,
            children: child.children || []
          })) || []
      }))
    })
  }
}

export const createUseEmittModuleMock = (
  overrides?: Partial<EmittEmitterMock>
) => {
  const emitter: EmittEmitterMock = {
    emit: vi.fn(),
    on: vi.fn(),
    off: vi.fn(),
    ...overrides
  }
  return {
    useEmitt: () => ({ emitter })
  }
}

export const createEventBusModuleMock = (overrides?: Partial<EventBusModuleMock>) => {
  return {
    default: {
      emit: vi.fn()
    },
    ...overrides
  }
}

export const createModelUtilModuleMock = (overrides?: Partial<ModelUtilModuleMock>) => {
  return {
    isDesktop: vi.fn(() => false),
    ...overrides
  }
}

export const createDatasetFormUtilModuleMock = (
  overrides?: Partial<DatasetFormUtilModuleMock>
) => {
  return {
    guid: vi.fn(() => 'test-guid-123'),
    ...overrides
  }
}

export const createDataVisualizationEditorUtilModuleMock = (
  overrides?: Partial<DataVisualizationEditorUtilModuleMock>
) => {
  return {
    DEFAULT_CANVAS_STYLE_DATA_DARK: {},
    DEFAULT_CANVAS_STYLE_DATA_LIGHT: {},
    DEFAULT_CANVAS_STYLE_DATA_SCREEN_DARK: {},
    ...overrides
  }
}

export const createPanelModuleMock = (overrides?: Partial<PanelModuleMock>) => {
  return {
    default: {},
    ...overrides
  }
}

export const createViewUtilsModuleMock = (overrides?: Partial<ViewUtilsModuleMock>) => {
  return {
    viewFieldTimeTrans: vi.fn(),
    ...overrides
  }
}

export const createComponentUtilsModuleMock = (overrides?: Partial<ComponentUtilsModuleMock>) => {
  return {
    filterEnumParams: vi.fn(),
    filterEnumParamsReduce: vi.fn(),
    filterParamsOptions: vi.fn(),
    ...overrides
  }
}

export const createEmbeddedApiModuleMock = (overrides?: Partial<EmbeddedApiModuleMock>) => {
  return {
    embeddedInitIframeApi: vi.fn().mockResolvedValue({ data: [] }),
    embeddedGetTokenArgsApi: vi.fn().mockResolvedValue({ data: { token: '', allowedOrigins: [] } }),
    ...overrides
  }
}

export const createDvMainStoreWithOutModuleMock = (
  overrides?: Partial<DvMainStoreWithOutMock>
) => {
  const store: DvMainStoreWithOutMock = {
    curOriginThemes: 'light',
    canvasAttachInfo: {},
    setComponentData: vi.fn(),
    setCanvasStyle: vi.fn(),
    updateCurDvInfo: vi.fn(),
    setCanvasViewInfo: vi.fn(),
    setNowPanelTrackInfo: vi.fn(),
    setNowPanelJumpInfo: vi.fn(),
    updateDvInfoCall: vi.fn(),
    removeGroupArea: vi.fn(),
    ...overrides
  }
  return {
    dvMainStoreWithOut: vi.fn(() => store)
  }
}

export const createSnapshotStoreWithOutModuleMock = (
  overrides?: Partial<SnapshotStoreWithOutMock>
) => {
  const store: SnapshotStoreWithOutMock = {
    resetStyleChangeTimes: vi.fn(),
    ...overrides
  }
  return {
    snapshotStoreWithOut: vi.fn(() => store)
  }
}

export const createAppearanceStoreWithOutModuleMock = (
  overrides?: Partial<AppearanceStoreWithOutMock>
) => {
  const store: AppearanceStoreWithOutMock = {
    setCurrentFont: vi.fn(),
    ...overrides
  }
  return {
    useAppearanceStoreWithOut: vi.fn(() => store)
  }
}

export const createAppStoreWithOutModuleMock = (overrides?: Partial<AppStoreWithOutMock>) => {
  const store: AppStoreWithOutMock = {
    getIsIframe: false,
    ...overrides
  }
  return {
    useAppStoreWithOut: vi.fn(() => store)
  }
}

export const createPermissionModuleMock = (overrides?: Partial<PermissionModuleMock>) => {
  const moduleMock: PermissionModuleMock = {
    pathValid: vi.fn().mockReturnValue(true),
    ...overrides
  }
  return moduleMock
}

type EmbeddedStoreMock = {
  allowedOrigins: string[]
  parent: boolean
  resourceId: string
  dvId: string
  chartId: string
  baseUrl: string
  setParam: ReturnType<typeof vi.fn>
  setResourceId: ReturnType<typeof vi.fn>
  setEmbedReady: ReturnType<typeof vi.fn>
  setJumpInfoParam: ReturnType<typeof vi.fn>
}

export const createEmbeddedStoreMock = (
  overrides?: Partial<EmbeddedStoreMock>
): EmbeddedStoreMock => {
  return {
    allowedOrigins: [],
    parent: false,
    resourceId: '',
    dvId: '',
    chartId: '',
    baseUrl: '',
    setParam: vi.fn(),
    setResourceId: vi.fn(),
    setEmbedReady: vi.fn(),
    setJumpInfoParam: vi.fn(),
    ...overrides
  }
}

export const createEmbeddedModuleMock = (store: EmbeddedStoreMock) => {
  return {
    useEmbedded: vi.fn(() => store)
  }
}

export const createEmbeddedUtilsModuleMock = (defaultAllowed = true) => {
  return {
    isAllowedEmbeddedMessageOrigin: vi.fn(() => defaultAllowed)
  }
}

export const useI18nModuleMock = {
  useI18n: () => ({
    t: (key: string) => key
  }),
  t: (key: string) => key
}

export const elementPlusSecondaryModuleMock = {
  ElMessage: {
    warning: vi.fn(),
    success: vi.fn(),
    error: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn().mockResolvedValue(undefined),
    close: vi.fn()
  }
}
