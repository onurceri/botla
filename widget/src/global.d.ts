declare module '*.css?raw' {
  const content: string
  export default content
}

// Window type extensions for widget globals
interface Window {
  __CBW_PARAMS?: Record<string, string>
  ChatbotWidget: {
    mount: () => void
    unmount: () => void
  }
  getCaptchaToken?: (siteKey: string) => Promise<string>
}
