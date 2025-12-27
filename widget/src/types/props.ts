import type { CustomBranding, WidgetPosition, PositionStrategy } from './index'

export interface WidgetThemeProps {
  themeColor?: string
  headerColor?: string
  headerTextColor?: string
  botMessageColor?: string
  botMessageTextColor?: string
  userMessageColor?: string
  userMessageTextColor?: string
  fontFamily?: string
  panelBg?: string
  chatBg?: string
  inputBg?: string
  inputText?: string
  bubbleRadius?: string
  sendButtonColor?: string
}

export interface WidgetLayoutProps {
  position?: WidgetPosition
  positionStrategy?: PositionStrategy
  panelHeight?: string
  panelWidth?: string
  previewMode?: boolean
}

export interface WidgetBrandingProps {
  hideBrandingOverride?: boolean
  customBrandingOverride?: CustomBranding
}

export interface WidgetAppProps extends 
  WidgetThemeProps, 
  WidgetLayoutProps, 
  WidgetBrandingProps {
  // Required
  chatbotId: string
  
  // API
  apiBase?: string
  embedTokenUrl?: string
  captchaSiteKey?: string
  
  // Bot customization
  botNameOverride?: string
  botIconOverride?: string
  welcome?: string
  suggestions?: string[]
  
  // Session
  resetSession?: boolean
  sessionIdOverride?: string
  
  // Behavior
  autoOpen?: boolean
  useOverrides?: boolean
  
  // Callbacks
  onOpenChange?: (isOpen: boolean) => void
}
