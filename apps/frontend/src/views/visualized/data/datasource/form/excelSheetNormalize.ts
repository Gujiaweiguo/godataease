export interface ExcelSheetField {
  originName?: string
  name?: string
  fieldIndex?: number
  checked?: boolean
}

export interface UploadedExcelSheet {
  fields?: ExcelSheetField[]
  data?: unknown[]
  jsonArray?: Record<string, unknown>[]
  [key: string]: unknown
}

export const isSheetEnabled = (sheet: { sheet?: boolean; isSheet?: boolean } | null | undefined) =>
  Boolean(sheet?.sheet || sheet?.isSheet)

export const buildSheetJsonArray = (sheet: UploadedExcelSheet): Record<string, unknown>[] => {
  if (Array.isArray(sheet?.jsonArray) && sheet.jsonArray.length > 0) {
    return sheet.jsonArray
  }

  const fields = Array.isArray(sheet?.fields) ? sheet.fields : []
  const dataRows = Array.isArray(sheet?.data) ? sheet.data : []

  return dataRows.map(row => {
    const rowValues = Array.isArray(row) ? row : []
    const item: Record<string, unknown> = {}

    fields.forEach((field, index) => {
      const key = String(field?.originName || field?.name || '').trim()
      if (!key) {
        return
      }

      const valueIndex = Number.isInteger(field?.fieldIndex) ? Number(field?.fieldIndex) : index
      item[key] = rowValues[valueIndex] ?? ''
    })

    return item
  })
}

export const normalizeUploadedSheets = <T extends UploadedExcelSheet = UploadedExcelSheet>(
  sheets: unknown[]
): T[] => {
  return sheets
    .filter(sheet => Boolean(sheet))
    .map(sheet => {
      const currentSheet: UploadedExcelSheet =
        typeof sheet === 'object' && sheet !== null ? (sheet as UploadedExcelSheet) : {}

      const normalizedFields: ExcelSheetField[] = Array.isArray(currentSheet.fields)
        ? currentSheet.fields.map(field => ({
            ...field,
            checked: field?.checked !== false
          }))
        : []

      return {
        ...currentSheet,
        fields: normalizedFields,
        jsonArray: buildSheetJsonArray({
          ...currentSheet,
          fields: normalizedFields
        })
      } as T
    })
}
