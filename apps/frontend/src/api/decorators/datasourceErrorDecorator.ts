/**
 * 数据源 API 错误处理装饰器
 * 为数据源 API 方法提供统一的错误处理
 */
export function withDatasourceError<T>(apiCall: () => Promise<T>): Promise<T> {
  return apiCall().catch((error: unknown) => {
    if (error instanceof Error) {
      throw error
    }
    throw new Error('Datasource request failed')
  })
}
