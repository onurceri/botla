import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import fs from 'fs'
import path from 'path'

test('Widget embed secure: auto-open shows input', async ({ page }) => {
  await page.route('http://api.test/api/v1/public/chatbots/bot1', async route => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({
      id: 'bot1', theme_color: '#3b82f6', welcome_message: 'Merhaba!', position: 'bottom-right',
      bot_message_color: '#3b82f6', user_message_color: '#f3f4f6', bot_message_text_color: '#ffffff', user_message_text_color: '#1f2937',
      chat_font_family: 'Inter, sans-serif', chat_header_color: '#3b82f6', chat_header_text_color: '#ffffff'
    }) })
  })
  await page.route('http://api.test/api/v1/public/chatbots/bot1/chat', async route => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ response: 'OK', tokens_used: 1, sources_used: [] }) })
  })

  const __filename = fileURLToPath(import.meta.url)
  const __dirname = path.dirname(__filename)
  const widgetPath = path.resolve(__dirname, '../../widget/dist/widget.js')
  const code = fs.readFileSync(widgetPath, 'utf-8')

  await page.goto('about:blank')
  await page.addScriptTag({ content: `window.getCaptchaToken = async () => 'CAP'` })
  await page.addScriptTag({ content: `window.__CBW_PARAMS={"chatbot-id":"bot1","api-base":"http://api.test","embed-token-url":"http://token.test/emit","captcha-site-key":"site","auto-open":"1"}` })
  await page.addScriptTag({ content: code })

  const host = await page.waitForSelector('#chatbot-widget-host', { state: 'attached' })
  await page.waitForFunction(() => {
    const h = document.getElementById('chatbot-widget-host')
    return !!h && !!h.shadowRoot && !!h.shadowRoot.querySelector('.cbw-input input')
  })
  const input = await host.evaluateHandle((el) => el.shadowRoot!.querySelector('.cbw-input input'))
  expect(await (input.asElement()!).isVisible()).toBeTruthy()
})
