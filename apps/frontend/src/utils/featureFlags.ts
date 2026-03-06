export const DYNAMIC_NAVIGATION_FLAG_KEY = 'feature.dynamic-navigation'

const FALSE_VALUES = ['0', 'false', 'off', 'disabled']

export const isDynamicNavigationEnabled = () => {
  if (typeof window === 'undefined') {
    return true
  }
  const rawValue = window.localStorage.getItem(DYNAMIC_NAVIGATION_FLAG_KEY)
  if (!rawValue) {
    return true
  }
  return !FALSE_VALUES.includes(rawValue.trim().toLowerCase())
}
