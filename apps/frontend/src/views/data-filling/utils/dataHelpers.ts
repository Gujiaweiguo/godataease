import type { FormFieldValue } from '@/views/data-filling/types'

type DatasourceOption = {
  value: number
  label: string
}

type BuiltInTableOption = {
  label: string
  value: string
}

export const isEmptyFieldValue = (value: FormFieldValue | undefined): boolean => {
  if (Array.isArray(value)) {
    return value.length === 0
  }
  return value === undefined || value === null || value === ''
}

export const isCancelableAction = (error: unknown): boolean => {
  return error === 'cancel' || error === 'close'
}

export const parseRouteFormId = (value: unknown): number | null => {
  const rawValue = Array.isArray(value) ? value[0] : value
  if (typeof rawValue !== 'string' || !rawValue.trim()) {
    return null
  }

  const parsedValue = Number(rawValue)
  return Number.isFinite(parsedValue) && parsedValue > 0 ? parsedValue : null
}

export const isBlankPayload = (payload: Record<string, FormFieldValue>): boolean => {
  const keys = Object.keys(payload).filter(key => key !== 'id')
  if (!keys.length) {
    return true
  }
  return keys.every(key => isEmptyFieldValue(payload[key]))
}

export const normalizeDatasourceOptions = (value: unknown): DatasourceOption[] => {
  if (!Array.isArray(value)) {
    return []
  }

  return value.reduce<DatasourceOption[]>((result, item) => {
    if (!item || typeof item !== 'object') {
      return result
    }

    const record = item as Record<string, unknown>
    const id = record.id ?? record.value
    const label = record.name ?? record.label
    if (typeof id === 'number' && Number.isFinite(id) && (typeof label === 'string' || typeof label === 'number')) {
      result.push({
        value: id,
        label: String(label)
      })
    }
    return result
  }, [])
}

export const normalizeBuiltInTableOptions = (value: unknown): BuiltInTableOption[] => {
  if (!Array.isArray(value)) {
    return []
  }

  return value.reduce<BuiltInTableOption[]>((result, item) => {
    if (typeof item === 'string') {
      result.push({ label: item, value: item })
      return result
    }

    if (!item || typeof item !== 'object') {
      return result
    }

    const record = item as Record<string, unknown>
    const candidate = record.tableName ?? record.name ?? record.label ?? record.value
    if (typeof candidate === 'string') {
      result.push({
        label: candidate,
        value: candidate
      })
    }
    return result
  }, [])
}
