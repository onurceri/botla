import { test, expect, type Page } from '@playwright/test'
import { readFileSync } from 'node:fs'
import path from 'node:path'

// Captures the README screenshots from the data that scripts/demo/seed.sh creates.
const dir = __dirname
const out = (name: string) => path.join(dir, '../../docs/screenshots', name)
const WIDGET_URL = process.env.WIDGET_URL || 'http://localhost:5174'
const EMAIL = process.env.DEMO_EMAIL || 'demo@example.com'
const PASSWORD = process.env.DEMO_PASSWORD || 'DemoPass123!'

// Hide toasts, dev overlays and animations so captures are stable.
const QUIET_CSS = `
  *, *::before, *::after { animation-duration: 0s !important; transition-duration: 0s !important; caret-color: transparent !important; }
  [data-sonner-toaster], [role="status"][aria-live], vite-error-overlay { display: none !important; }
`

// SHOTS=03,04 retakes only the screenshots whose names start with those prefixes.
const ONLY = process.env.SHOTS?.split(',').filter(Boolean)

async function shoot(page: Page, name: string) {
  if (ONLY && !ONLY.some((p) => name.startsWith(p))) return
  await page.addStyleTag({ content: QUIET_CSS })
  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(800)
  await page.screenshot({ path: out(name) })
}

test('readme screenshots', async ({ page }) => {
  // Login and the API are rate-limited; retry while a burst (e.g. the seed run) cools down.
  await expect(async () => {
    await page.goto('/login')
    await page.getByTestId('login-page-email-input').fill(EMAIL)
    await page.getByTestId('login-page-password-input').fill(PASSWORD)
    await page.getByTestId('login-page-submit-button').click()
    await page.waitForURL(/\/dashboard/, { timeout: 5_000 })
  }).toPass({ timeout: 120_000, intervals: [10_000] })

  // 1. Bot list
  await expect(async () => {
    await page.goto('/dashboard/chatbots')
    await expect(page.getByText('Halvane Support')).toBeVisible({ timeout: 5_000 })
  }).toPass({ timeout: 90_000, intervals: [5_000] })
  await shoot(page, '01-chatbots.png')

  await page.getByTestId('chatbot-card').filter({ hasText: 'Halvane Support' }).getByTestId('chatbot-manage-button').click()
  await page.waitForURL(/\/chatbots\/[0-9a-f-]{36}/)
  const botId = page.url().match(/chatbots\/([0-9a-f-]{36})/)![1]
  const bot = (tab: string) => `/dashboard/chatbots/${botId}/${tab}`

  // 2. Sources with ingestion status
  await page.goto(bot('sources'))
  await shoot(page, '02-sources.png')

  // 3. Appearance settings next to the live Playground preview
  await page.goto(bot('playground'))
  await page.waitForTimeout(2500) // widget iframe boots
  await shoot(page, '03-playground.png')

  // 5. Analytics (scrolled so the trend chart is in view) and 6. handoff requests
  await page.goto(bot('insights'))
  await page.locator('.recharts-surface').first().waitFor()
  await page.getByText('Genel Bakış').scrollIntoViewIfNeeded()
  await page.getByText('Genel Bakış').evaluate((el) => el.scrollIntoView({ block: 'start' }))
  await shoot(page, '05-analytics.png')
  await page.getByText('Raporlar', { exact: true }).last().evaluate((el) => el.scrollIntoView({ block: 'start' }))
  await page.getByRole('tab', { name: 'Destek Talepleri' }).click()
  await shoot(page, '06-handoff.png')

  // 7. Deploy tab; blank any embed secret before the capture
  await page.goto(bot('deploy'))
  await page.waitForLoadState('networkidle')
  await page.evaluate(() => {
    document.querySelectorAll<HTMLInputElement>('input').forEach((i) => {
      if (/secret/i.test(i.closest('div')?.textContent || '') && i.value) i.value = '••••••••••••••••'
    })
  })
  await shoot(page, '07-deploy.png')

  // 4. The widget on a sample customer page, answering a real multi-part question
  const html = readFileSync(path.join(dir, 'sample-page.html'), 'utf8')
    .replace('__WIDGET_SRC__', `${WIDGET_URL}/src/widget.tsx`)
    .replace('__BOT_ID__', botId)
  await page.route(`${WIDGET_URL}/halvane-demo.html`, (r) =>
    r.fulfill({ contentType: 'text/html', body: html }),
  )
  await page.goto(`${WIDGET_URL}/halvane-demo.html`)
  await page.locator('.cbw-bubble').waitFor()
  // The widget lives in a shadow root, so stop its animations there too.
  await page.evaluate((css) => {
    const style = document.createElement('style')
    style.textContent = css
    document.getElementById('chatbot-widget-host')?.shadowRoot?.appendChild(style)
  }, QUIET_CSS)
  await page.locator('.cbw-bubble').click({ force: true })
  const input = page.locator('.cbw-panel textarea, .cbw-panel input[type="text"]').first()
  await input.fill('How long does the Beacon S2 battery last, and does it work on 5 GHz Wi-Fi?')
  await input.press('Enter')
  await expect(page.locator('.cbw-msg-row.assistant')).toHaveCount(2, { timeout: 60_000 })
  await page.waitForTimeout(1500)
  await shoot(page, '04-widget.png')
})
