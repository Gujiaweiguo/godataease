import { ref, onMounted, onBeforeUnmount, Ref } from 'vue'
import { useEmbedded } from '@/store/modules/embedded'
import { isAllowedEmbeddedMessageOrigin } from '@/utils/embedded'
import type {
  EmbeddingEventType,
  EmbeddingEvent,
  ParamUpdatePayload,
  InteractionPayload,
  InitReadyPayload,
  ReadyPayload,
  ErrorPayload,
  DeInitPayload,
  CanvasInitPayload,
  AttachParamsPayload,
  JumpToTargetPayload,
  ModuleInitPayload,
  ModuleUpdatePayload,
  ModuleInteractionPayload,
  EmbeddingEventPayload
} from '@/events/embedding/types'
import type { PayloadForEventType } from '@/events/embedding/payloads'
import { validatePayload } from '@/events/embedding/payloads'

/**
 * Parent frame communication composable for embedded mode.
 *
 * Handles bidirectional communication between parent frame (host page)
 * and child frame (embedded content) using standardized event types.
 */
export function useEmbeddedParentCommunication() {
  const store = useEmbedded()
  const messageHandler: Ref<((event: MessageEvent) => void) | null> = ref(null)

  /**
   * Parse and validate incoming message from child frame.
   */
  const parseIncomingMessage = (event: MessageEvent): EmbeddingEvent | null => {
    try {
      if (typeof event.data === 'string') {
        const parsed = JSON.parse(event.data)

        if (parsed.type && isValidEmbeddingEventType(parsed.type)) {
          const eventType = parsed.type as EmbeddingEventType
          if (validatePayload(eventType, parsed.payload)) {
            return {
              type: eventType,
              payload: parsed.payload,
              source: 'child'
            }
          }
        }
      }
      return null
    } catch (error) {
      console.error('Failed to parse incoming message:', error)
      return null
    }
  }

  /**
   * Validate event type.
   */
  const isValidEmbeddingEventType = (value: string): boolean => {
    const validTypes: string[] = [
      'param_update',
      'user_interaction',
      'init_ready',
      'ready',
      'error',
      'de_init',
      'canvas_init',
      'attach_params',
      'jump_to_target',
      'module_init',
      'module_update',
      'module_interaction'
    ]
    return validTypes.includes(value)
  }

  /**
   * Listen for messages from child frame.
   * Handles all standardized embedding events.
   */
  const listenForChildMessages = () => {
    const handler = (event: MessageEvent) => {
      if (!isAllowedEmbeddedMessageOrigin(event.origin, store.allowedOrigins, true)) {
        console.warn(`Message from untrusted origin blocked: ${event.origin}`)
        return
      }

      const embeddingEvent = parseIncomingMessage(event)
      if (!embeddingEvent) {
        return
      }

      handleChildEvent(embeddingEvent)
    }

    messageHandler.value = handler
    window.addEventListener('message', handler)
  }

  /**
   * Handle events received from child frame.
   * Dispatches to appropriate store actions.
   */
  const handleChildEvent = (event: EmbeddingEvent): void => {
    switch (event.type) {
      case 'param_update':
        handleParamUpdate(event.payload as ParamUpdatePayload)
        break

      case 'user_interaction':
        handleInteraction(event.payload as InteractionPayload)
        break

      case 'init_ready':
        handleInitReady(event.payload as InitReadyPayload)
        break

      case 'ready':
        handleReady(event.payload as ReadyPayload)
        break

      case 'error':
        handleError(event.payload as ErrorPayload)
        break

      case 'de_init':
        handleDeInit(event.payload as DeInitPayload)
        break

      case 'canvas_init':
        handleCanvasInit(event.payload as CanvasInitPayload)
        break

      case 'attach_params':
        handleAttachParams(event.payload as AttachParamsPayload)
        break

      case 'jump_to_target':
        handleJumpToTarget(event.payload as JumpToTargetPayload)
        break

      case 'module_init':
        handleModuleInit(event.payload as ModuleInitPayload)
        break

      case 'module_update':
        handleModuleUpdate(event.payload as ModuleUpdatePayload)
        break

      case 'module_interaction':
        handleModuleInteraction(event.payload as ModuleInteractionPayload)
        break

      default:
        console.warn('Unknown event type:', event.type)
    }
  }

  /**
   * Handle parameter update from child.
   */
  const handleParamUpdate = (payload: ParamUpdatePayload): void => {
    if (payload.resourceId && payload.resourceId === store.resourceId) {
      store.setResourceId(payload.resourceId)

      Object.keys(payload).forEach(key => {
        if (key !== 'resourceId' && key !== 'timestamp') {
          const value = (payload as any)[key]
          store.setParam(key, value)
        }
      })
    }
  }

  /**
   * Handle user interaction from child.
   */
  const handleInteraction = (payload: InteractionPayload): void => {
    if (payload.param && payload.value !== undefined) {
      store.setParam(payload.param, payload.value)
    }

    console.debug('User interaction:', payload)
  }

  /**
   * Handle init ready event from child.
   */
  const handleInitReady = (payload: InitReadyPayload): void => {
    console.log('Child frame ready for initialization:', payload.resourceId)

    if (payload.resourceId) {
      store.setResourceId(payload.resourceId)
    }
  }

  /**
   * Handle ready event from child.
   */
  const handleReady = (payload: ReadyPayload): void => {
    console.log('Child frame fully ready:', payload.component || 'unknown')
    store.setEmbedReady(true)
  }

  /**
   * Handle error from child.
   */
  const handleError = (payload: ErrorPayload): void => {
    console.error('Child frame error:', payload.message, payload.context)

    if (store.parent && window.parent !== window.top) {
      const errorMessage = `${payload.message}${payload.context ? ` (${payload.context})` : ''}`
      const targetPm = `dataease-embedded-host-error:${errorMessage}`
      window.parent.postMessage(targetPm, '*')
    }
  }

  /**
   * Handle de-init event from child.
   */
  const handleDeInit = (payload: DeInitPayload): void => {
    console.log('Child frame de-initializing:', payload.reason)

    if (payload.params) {
      Object.keys(payload.params).forEach(key => {
        store.setParam(key, null)
      })
    }
  }

  /**
   * Handle canvas init event from child.
   */
  const handleCanvasInit = (payload: CanvasInitPayload): void => {
    console.log('Canvas initialized:', payload.dimensions)
  }

  /**
   * Handle attach params event from child.
   * For DIV-based embed mode where child needs inner params.
   */
  const handleAttachParams = (payload: AttachParamsPayload): void => {
    if (payload.dvId && payload.dvId === store.dvId) {
      if (payload.innerParams) {
        Object.keys(payload.innerParams).forEach(key => {
          store.setParam(key, payload.innerParams[key])
        })
      }
    }
  }

  /**
   * Handle jump to target event from child.
   */
  const handleJumpToTarget = (payload: JumpToTargetPayload): void => {
    if (
      (payload.chartId && payload.chartId === store.chartId) ||
      (payload.dvId && payload.dvId === store.dvId)
    ) {
      if (payload.jumpInfoParam) {
        store.setJumpInfoParam(payload.jumpInfoParam)
      }

      if (payload.targetUrl) {
        console.log('Jump to target:', payload.targetUrl)
      }
    }
  }

  /**
   * Handle module init event from child.
   */
  const handleModuleInit = (payload: ModuleInitPayload): void => {
    console.log('Module initialized:', payload.moduleId, payload.moduleName)

    if (payload.params) {
      Object.keys(payload.params).forEach(key => {
        store.setParam(key, payload.params[key])
      })
    }
  }

  /**
   * Handle module update event from child.
   */
  const handleModuleUpdate = (payload: ModuleUpdatePayload): void => {
    console.log('Module updated:', payload.moduleId)

    if (payload.params) {
      Object.keys(payload.params).forEach(key => {
        store.setParam(key, payload.params[key])
      })
    }

    if (payload.changes) {
      payload.changes.forEach(change => {
        store.setParam(change.key, change.newValue)
      })
    }
  }

  /**
   * Handle module interaction event from child.
   */
  const handleModuleInteraction = (payload: ModuleInteractionPayload): void => {
    console.debug('Module interaction:', payload.interactionType, payload.target)

    if (payload.data) {
      Object.keys(payload.data).forEach(key => {
        store.setParam(key, payload.data[key])
      })
    }
  }

  /**
   * Emit event to child frame.
   * Posts standardized event to child frame.
   */
  const emitToChild = <T extends EmbeddingEventType>(
    type: T,
    payload: PayloadForEventType<T>
  ): void => {
    if (!store.parent || window.parent === window.top) {
      console.warn('Not in embedded mode, skipping emit to child')
      return
    }

    const event: EmbeddingEvent<PayloadForEventType<T>> = {
      type,
      payload: payload as EmbeddingEventPayload,
      source: 'parent'
    }

    const targetPm = `dataease-embedded-host:${JSON.stringify(event)}`
    window.parent.postMessage(targetPm, '*')
  }

  /**
   * Post legacy format message to child.
   * Maintains backward compatibility with existing code.
   */
  const postLegacyMessage = (message: string): void => {
    if (!store.parent || window.parent === window.top) {
      return
    }

    const targetPm = `dataease-embedded-host${message}`
    window.parent.postMessage(targetPm, '*')
  }

  /**
   * Clean up message listener on unmount.
   */
  onMounted(() => {
    listenForChildMessages()
  })

  onBeforeUnmount(() => {
    if (messageHandler.value) {
      window.removeEventListener('message', messageHandler.value)
      messageHandler.value = null
    }
  })

  return {
    listenForChildMessages,
    emitToChild,
    postLegacyMessage
  }
}
