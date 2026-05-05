import type {
  DataFillingFieldType,
  DataFillingFormSchema,
  FormFieldConfig,
  FormFieldOption,
  FormFieldValue
} from '@/views/data-filling/types'

type BackendFieldMapping = Record<string, unknown>
type BackendFieldSettings = Record<string, unknown>

const FIELD_TYPES: ReadonlyArray<DataFillingFieldType> = ['text', 'number', 'decimal', 'date', 'datetime', 'select']

export const isRecord = (value: unknown): value is Record<string, unknown> => {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

export const normalizeOptions = (value: unknown): FormFieldOption[] => {
  if (!Array.isArray(value)) {
    return []
  }

  return value.reduce<FormFieldOption[]>((result, item) => {
    if (isRecord(item)) {
      const optionName = item.name ?? item.label ?? item.value
      const optionValue = item.value ?? item.name ?? item.label

      if (optionName != null && optionValue != null) {
        result.push({
          name: String(optionName),
          value: String(optionValue),
          disabled: Boolean(item.disabled),
          description: item.description ? String(item.description) : undefined
        })
      }

      return result
    }

    if (typeof item === 'string' || typeof item === 'number') {
      result.push({
        name: String(item),
        value: String(item)
      })
    }

    return result
  }, [])
}

export const normalizeFieldType = (value: unknown): string => {
  return typeof value === 'string' && value.trim() ? value.trim().toLowerCase() : 'text'
}

export const resolveFieldTypeFromMapping = (value: unknown): string | undefined => {
  const mappingType = normalizeFieldType(value)

  if (mappingType === 'nvarchar') {
    return 'text'
  }

  return FIELD_TYPES.includes(mappingType) ? mappingType : undefined
}

const normalizeDefaultValue = (value: unknown): FormFieldValue | undefined => {
  if (
    typeof value === 'string' ||
    typeof value === 'number' ||
    typeof value === 'boolean' ||
    value == null ||
    Array.isArray(value) ||
    isRecord(value)
  ) {
    return value as FormFieldValue
  }

  return undefined
}

export const toFieldConfig = (value: unknown, index: number): FormFieldConfig | null => {
  if (!isRecord(value)) {
    return null
  }

  const settings = isRecord(value.settings) ? (value.settings as BackendFieldSettings) : undefined
  const mapping = settings && isRecord(settings.mapping) ? (settings.mapping as BackendFieldMapping) : undefined

  const fieldName = typeof mapping?.columnName === 'string' && mapping.columnName.trim() ? mapping.columnName.trim() : undefined
  const legacyField = typeof value.field === 'string' && value.field.trim() ? value.field.trim() : undefined
  const legacyName = typeof value.name === 'string' && value.name.trim() ? value.name.trim() : undefined
  const labelFromSettings = typeof settings?.name === 'string' && settings.name.trim() ? settings.name.trim() : undefined
  const legacyLabel = typeof value.label === 'string' && value.label.trim() ? value.label.trim() : undefined

  const name = fieldName ?? legacyName ?? legacyField ?? `field_${index + 1}`
  const label = labelFromSettings ?? legacyLabel ?? legacyName ?? `Field ${index + 1}`
  const mappingType = resolveFieldTypeFromMapping(mapping?.type)
  const explicitType = resolveFieldTypeFromMapping(value.type ?? value.fieldType) ?? normalizeFieldType(value.type ?? value.fieldType)

  return {
    id: typeof value.id === 'string' || typeof value.id === 'number' ? value.id : undefined,
    field: fieldName ?? legacyField ?? name,
    name,
    label,
    type: mappingType ?? explicitType,
    required: settings ? Boolean(settings.required) : Boolean(value.required),
    placeholder:
      typeof settings?.placeholder === 'string'
        ? settings.placeholder
        : typeof value.placeholder === 'string'
        ? value.placeholder
        : undefined,
    defaultValue: normalizeDefaultValue(value.defaultValue),
    order: typeof value.order === 'number' ? value.order : index,
    options: normalizeOptions(settings?.options ?? value.options ?? value.optionList),
    optionDatasource:
      settings?.optionDatasource != null
        ? String(settings.optionDatasource)
        : value.optionDatasource
        ? String(value.optionDatasource)
        : undefined,
    optionTable:
      typeof settings?.optionTable === 'string'
        ? settings.optionTable
        : typeof value.optionTable === 'string'
        ? value.optionTable
        : undefined,
    optionColumn:
      typeof settings?.optionColumn === 'string'
        ? settings.optionColumn
        : typeof value.optionColumn === 'string'
        ? value.optionColumn
        : undefined,
    optionOrder:
      typeof settings?.optionOrder === 'string'
        ? settings.optionOrder
        : typeof value.optionOrder === 'string'
        ? value.optionOrder
        : undefined,
    precision:
      typeof mapping?.accuracy === 'number'
        ? mapping.accuracy
        : typeof value.precision === 'number'
        ? value.precision
        : undefined,
    format: value.format ? String(value.format) : undefined,
    multiple: settings ? Boolean(settings.multiple) : Boolean(value.multiple),
    extra: isRecord(value.extra) ? value.extra : undefined
  }
}

export const parseFormSchema = (forms: string): DataFillingFormSchema => {
  if (!forms.trim()) {
    return []
  }

  try {
    const parsed = JSON.parse(forms) as unknown
    const source = Array.isArray(parsed)
      ? parsed
      : isRecord(parsed) && Array.isArray(parsed.fields)
      ? parsed.fields
      : isRecord(parsed) && Array.isArray(parsed.forms)
      ? parsed.forms
      : []

    return source
      .map((item, index) => toFieldConfig(item, index))
      .filter((item): item is FormFieldConfig => item !== null)
      .sort((prev, next) => (prev.order ?? 0) - (next.order ?? 0))
  } catch {
    return []
  }
}

export const resolveFieldKey = (field: FormFieldConfig, index: number): string => {
  return field.field || field.name || `field_${index + 1}`
}

export const resolveComponentType = (
  type: string
): 'input' | 'number' | 'decimal' | 'date' | 'datetime' | 'select' => {
  const normalizedType = normalizeFieldType(type)

  if (normalizedType === 'number') {
    return 'number'
  }
  if (normalizedType === 'decimal') {
    return 'decimal'
  }
  if (normalizedType === 'date') {
    return 'date'
  }
  if (normalizedType === 'datetime') {
    return 'datetime'
  }
  if (normalizedType === 'select') {
    return 'select'
  }

  return 'input'
}

export const buildPlaceholder = (field: FormFieldConfig, componentType: string): string => {
  if (field.placeholder) {
    return field.placeholder
  }

  if (componentType === 'select' || componentType === 'date' || componentType === 'datetime') {
    return `请选择${field.label}`
  }

  return `请输入${field.label}`
}
