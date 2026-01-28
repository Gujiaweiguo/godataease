/**
 * Payload interfaces for embedding event communication.
 *
 * Defines TypeScript interfaces for all standardized event payloads used
 * in parent-child communication for embedding scenarios.
 */

import type { EmbeddingEventType } from './types'

/**
 * Base interface for all embedding event payloads.
 */
interface BaseEmbeddingPayload {
  /**
   * Timestamp when the event was created.
   */
  timestamp?: number
}

/**
 * Payload for PARAM_UPDATE event.
 * Notifies parent of parameter updates from child frame.
 */
export interface ParamUpdatePayload extends BaseEmbeddingPayload {
  /**
   * Unique identifier for the resource (dashboard, screen, chart, etc.).
   */
  resourceId?: string

  /**
   * Parameter key-value pairs.
   * Dynamic structure allows any parameter name.
   */
  [key: string]: any
}

/**
 * Payload for INTERACTION event.
 * Notifies parent of user interactions within embedded content.
 */
export interface InteractionPayload extends BaseEmbeddingPayload {
  /**
   * Type of interaction (click, select, filter, etc.).
   */
  interactionType?: string

  /**
   * Parameter name affected by interaction.
   */
  param?: string

  /**
   * New value for the parameter.
   */
  value?: any

  /**
   * Additional metadata about the interaction.
   */
  metadata?: Record<string, any>
}

/**
 * Payload for INIT_READY event.
 * Signals that canvas/screen is ready for initialization.
 */
export interface InitReadyPayload extends BaseEmbeddingPayload {
  /**
   * Unique identifier for the resource being embedded.
   */
  resourceId: string
}

/**
 * Payload for READY event.
 * Signals that embedded content is fully ready and interactive.
 */
export interface ReadyPayload extends BaseEmbeddingPayload {
  /**
   * Component type that is ready (dashboard, screen, chart, etc.).
   */
  component?: string

  /**
   * Unique identifier for the resource.
   */
  resourceId?: string
}

/**
 * Payload for ERROR event.
 * Signals an error occurred in embedded content.
 */
export interface ErrorPayload extends BaseEmbeddingPayload {
  /**
   * Error message or code.
   */
  message: string

  /**
   * Error code (optional, for programmatic error handling).
   */
  code?: string

  /**
   * Stack trace or additional error details.
   */
  details?: any

  /**
   * Context where error occurred (component, action, etc.).
   */
  context?: string
}

/**
 * Payload for DE_INIT event.
 * Signals de-initialization or cleanup is needed.
 */
export interface DeInitPayload extends BaseEmbeddingPayload {
  /**
   * Reason for de-initialization.
   */
  reason?: string

  /**
   * Parameters to reset or clean up.
   */
  params?: Record<string, any>
}

/**
 * Payload for CANVAS_INIT event.
 * Signals canvas is initialized and ready.
 */
export interface CanvasInitPayload extends BaseEmbeddingPayload {
  /**
   * Canvas dimensions.
   */
  dimensions?: {
    width: number
    height: number
  }

  /**
   * Canvas configuration options.
   */
  config?: Record<string, any>
}

/**
 * Payload for ATTACH_PARAMS event.
 * Signals inner parameters for DIV-based embed mode.
 */
export interface AttachParamsPayload extends BaseEmbeddingPayload {
  /**
   * Data visualization ID (for screens/datav).
   */
  dvId?: string

  /**
   * Inner parameters to attach.
   */
  innerParams?: Record<string, any>
}

/**
 * Payload for JUMP_TO_TARGET event.
 * Signals a jump/redirect action to a specific target.
 */
export interface JumpToTargetPayload extends BaseEmbeddingPayload {
  /**
   * Target chart ID.
   */
  chartId?: string

  /**
   * Data visualization ID.
   */
  dvId?: string

  /**
   * Jump information parameters.
   */
  jumpInfoParam?: Record<string, any>

  /**
   * Target URL or route.
   */
  targetUrl?: string

  /**
   * Jump type (chart-to-chart, screen-to-screen, etc.).
   */
  jumpType?: string
}

/**
 * Payload for MODULE_INIT event.
 * Signals module-level page is initialized.
 */
export interface ModuleInitPayload extends BaseEmbeddingPayload {
  /**
   * Module identifier.
   */
  moduleId?: string

  /**
   * Module name or path.
   */
  moduleName?: string

  /**
   * Initial module parameters.
   */
  params?: Record<string, any>
}

/**
 * Payload for MODULE_UPDATE event.
 * Signals module-level page parameters are updated.
 */
export interface ModuleUpdatePayload extends BaseEmbeddingPayload {
  /**
   * Module identifier.
   */
  moduleId?: string

  /**
   * Updated parameters.
   */
  params?: Record<string, any>

  /**
   * Parameter changes (before/after values).
   */
  changes?: Array<{
    key: string
    oldValue: any
    newValue: any
  }>
}

/**
 * Payload for MODULE_INTERACTION event.
 * Signals user interaction within module-level page.
 */
export interface ModuleInteractionPayload extends BaseEmbeddingPayload {
  /**
   * Module identifier.
   */
  moduleId?: string

  /**
   * Interaction type.
   */
  interactionType?: string

  /**
   * Affected parameter or element.
   */
  target?: string

  /**
   * Interaction data.
   */
  data?: Record<string, any>
}

/**
 * Union type of all valid embedding event payloads.
 * Used for type-safe event handling.
 */
export type EmbeddingEventPayload =
  | ParamUpdatePayload
  | InteractionPayload
  | InitReadyPayload
  | ReadyPayload
  | ErrorPayload
  | DeInitPayload
  | CanvasInitPayload
  | AttachParamsPayload
  | JumpToTargetPayload
  | ModuleInitPayload
  | ModuleUpdatePayload
  | ModuleInteractionPayload

/**
 * Complete event structure with type and payload.
 */
export interface EmbeddingEvent<T extends EmbeddingEventPayload = EmbeddingEventPayload> {
  /**
   * Event type from EmbeddingEventType enum.
   */
  type: EmbeddingEventType

  /**
   * Event payload.
   */
  payload: T

  /**
   * Event source (parent or child frame).
   */
  source?: 'parent' | 'child'
}

/**
 * Maps event types to their corresponding payload types.
 */
export type PayloadForEventType<T extends EmbeddingEventType> = T extends EmbeddingEventType.PARAM_UPDATE
  ? ParamUpdatePayload
  : T extends EmbeddingEventType.INTERACTION
    ? InteractionPayload
    : T extends EmbeddingEventType.INIT_READY
      ? InitReadyPayload
      : T extends EmbeddingEventType.READY
        ? ReadyPayload
        : T extends EmbeddingEventType.ERROR
          ? ErrorPayload
          : T extends EmbeddingEventType.DE_INIT
            ? DeInitPayload
            : T extends EmbeddingEventType.CANVAS_INIT
              ? CanvasInitPayload
              : T extends EmbeddingEventType.ATTACH_PARAMS
                ? AttachParamsPayload
                : T extends EmbeddingEventType.JUMP_TO_TARGET
                  ? JumpToTargetPayload
                  : T extends EmbeddingEventType.MODULE_INIT
                    ? ModuleInitPayload
                    : T extends EmbeddingEventType.MODULE_UPDATE
                      ? ModuleUpdatePayload
                      : T extends EmbeddingEventType.MODULE_INTERACTION
                        ? ModuleInteractionPayload
                        : never

/**
 * Validates if a payload matches expected structure for an event type.
 */
export const validatePayload = (
  eventType: EmbeddingEventType,
  payload: unknown
): payload is EmbeddingEventPayload => {
  if (!payload || typeof payload !== 'object') {
    return false
  }

  const p = payload as Partial<EmbeddingEventPayload>

  switch (eventType) {
    case EmbeddingEventType.INIT_READY:
      return typeof p.resourceId === 'string'
    case EmbeddingEventType.ERROR:
      return typeof p.message === 'string'
    case EmbeddingEventType.READY:
      return true // No required fields
    default:
      return true // Other events have no required fields
  }
}
