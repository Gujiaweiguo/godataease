/**
 * Event type registry for bidirectional parameter passing in embedded mode.
 *
 * This file defines the standardized event types used for parent-child
 * communication between embedded content and host pages.
 */

/**
 * Standardized event types for embedding communication.
 *
 * These events are used to coordinate parameter passing, lifecycle events,
 * and user interactions between parent frames (host) and child frames (embedded content).
 */
export enum EmbeddingEventType {
  // ========== Parameter Update Events ==========

  /**
   * Notifies parent of parameter updates from child.
   * Sent when user interaction changes query parameters or filters.
   */
  PARAM_UPDATE = 'param_update',

  /**
   * Notifies parent of user interactions within the embedded content.
   * Used to track user actions like clicks, selections, or navigation.
   */
  INTERACTION = 'user_interaction',

  // ========== Lifecycle Events ==========

  /**
   * Signals that the canvas/screen is ready for initialization.
   * Parent can send initialization parameters after receiving this event.
   */
  INIT_READY = 'init_ready',

  /**
   * Signals that the embedded content is fully ready and interactive.
   * Sent after all components are mounted and initialized.
   */
  READY = 'ready',

  /**
   * Signals an error occurred in the embedded content.
   * Parent should handle or display the error appropriately.
   */
  ERROR = 'error',

  /**
   * Signals de-initialization or cleanup is needed.
   * Used when embedded content is being unloaded or parameters are being reset.
   */
  DE_INIT = 'de_init',

  /**
   * Signals canvas is initialized and ready.
   * Specific to canvas-based embed modes (screens/datav).
   */
  CANVAS_INIT = 'canvas_init',

  /**
   * Signals inner parameters for DIV-based embed mode.
   * Used when parent frame provides initialization parameters to child frame.
   */
  ATTACH_PARAMS = 'attach_params',

  /**
   * Signals a jump/redirect action to a specific target.
   * Used for navigation between charts or screens.
   */
  JUMP_TO_TARGET = 'jump_to_target',

  // ========== Module-Level Page Events ==========

  /**
   * Signals module-level page is initialized.
   * Used for tree navigation and module-based embedding.
   */
  MODULE_INIT = 'module_init',

  /**
   * Signals module-level page parameters are updated.
   * Used when module tree navigation or filters change.
   */
  MODULE_UPDATE = 'module_update',

  /**
   * Signals user interaction within module-level page.
   * Used to track module-specific user actions.
   */
  MODULE_INTERACTION = 'module_interaction'
}

/**
 * Event direction flags for type safety.
 */
export type EventDirection = 'parent-to-child' | 'child-to-parent' | 'bidirectional'

/**
 * Mapping of event types to their typical direction.
 */
export const EventDirectionMap: Record<EmbeddingEventType, EventDirection> = {
  [EmbeddingEventType.PARAM_UPDATE]: 'child-to-parent',
  [EmbeddingEventType.INTERACTION]: 'child-to-parent',
  [EmbeddingEventType.INIT_READY]: 'child-to-parent',
  [EmbeddingEventType.READY]: 'child-to-parent',
  [EmbeddingEventType.ERROR]: 'child-to-parent',
  [EmbeddingEventType.DE_INIT]: 'child-to-parent',
  [EmbeddingEventType.CANVAS_INIT]: 'child-to-parent',
  [EmbeddingEventType.ATTACH_PARAMS]: 'parent-to-child',
  [EmbeddingEventType.JUMP_TO_TARGET]: 'child-to-parent',
  [EmbeddingEventType.MODULE_INIT]: 'child-to-parent',
  [EmbeddingEventType.MODULE_UPDATE]: 'child-to-parent',
  [EmbeddingEventType.MODULE_INTERACTION]: 'child-to-parent'
}

/**
 * Validates if an event type is recognized.
 */
export const isValidEmbeddingEventType = (value: string): value is EmbeddingEventType => {
  return Object.values(EmbeddingEventType).includes(value as EmbeddingEventType)
}
