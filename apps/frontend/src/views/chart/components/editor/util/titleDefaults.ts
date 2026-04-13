export const DEFAULT_TITLE_STYLE_BASE = {
  show: true,
  fontSize: 16,
  hPosition: 'left',
  vPosition: 'top',
  isItalic: false,
  isBolder: true,
  remarkShow: false,
  remark: '',
  fontFamily: '',
  letterSpace: '0',
  fontShadow: false,
  color: '',
  remarkBackgroundColor: ''
}

export const DEFAULT_TITLE_STYLE_LIGHT = {
  ...DEFAULT_TITLE_STYLE_BASE,
  color: '#000000',
  remarkBackgroundColor: '#ffffff'
}

export const DEFAULT_TITLE_STYLE_DARK = {
  ...DEFAULT_TITLE_STYLE_BASE,
  color: '#FFFFFF',
  remarkBackgroundColor: '#5A5C62'
}
