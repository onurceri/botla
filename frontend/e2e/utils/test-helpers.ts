/**
 * Test Helper Utilities
 * 
 * This module provides convenient helper functions for element selection,
 * selector generation, and common test operations following the established
 * naming conventions.
 */

import { Locator, Page, expect } from '@playwright/test'
import { SELECTORS, SelectorCategory } from './selectors'

// ============================================================================
// SELECTOR GENERATORS
// ============================================================================

/**
 * Generate a data-testid selector string
 */
export function testId(id: string): string {
  return `[data-testid="${id}"]`
}

/**
 * Get a locator by test ID
 */
export function getByTestId(pageOrLocator: Page | Locator, testId: string): Locator {
  return pageOrLocator.locator(`[data-testid="${testId}"]`)
}

/**
 * Get a button locator by its action and target
 */
export function getButton(
  pageOrLocator: Page | Locator,
  action: string,
  target?: string
): Locator {
  const id = target ? `btn-${action}-${target}` : `btn-${action}`
  return getByTestId(pageOrLocator, id)
}

/**
 * Get an input locator by its field name
 */
export function getInput(
  pageOrLocator: Page | Locator,
  fieldName: string
): Locator {
  return getByTestId(pageOrLocator, `input-${fieldName}`)
}

/**
 * Get a link locator by its destination
 */
export function getLink(
  pageOrLocator: Page | Locator,
  destination: string
): Locator {
  return getByTestId(pageOrLocator, `link-${destination}`)
}

/**
 * Get a tab locator by its name
 */
export function getTab(
  pageOrLocator: Page | Locator,
  tabName: string
): Locator {
  return getByTestId(pageOrLocator, `tab-${tabName}`)
}

/**
 * Get a modal locator by its purpose
 */
export function getModal(
  pageOrLocator: Page | Locator,
  purpose: string
): Locator {
  return getByTestId(pageOrLocator, `modal-${purpose}`)
}

/**
 * Get a card locator by its content type
 */
export function getCard(
  pageOrLocator: Page | Locator,
  contentType: string
): Locator {
  return getByTestId(pageOrLocator, `card-${contentType}`)
}

/**
 * Get a list locator by its content type
 */
export function getList(
  pageOrLocator: Page | Locator,
  contentType: string
): Locator {
  return getByTestId(pageOrLocator, `list-${contentType}`)
}

/**
 * Get an error message locator by context
 */
export function getError(
  pageOrLocator: Page | Locator,
  context: string
): Locator {
  return getByTestId(pageOrLocator, `error-${context}`)
}

/**
 * Get a success message locator by context
 */
export function getSuccess(
  pageOrLocator: Page | Locator,
  context: string
): Locator {
  return getByTestId(pageOrLocator, `success-${context}`)
}

/**
 * Get a loading indicator locator by context
 */
export function getLoading(
  pageOrLocator: Page | Locator,
  context: string
): Locator {
  return getByTestId(pageOrLocator, `loading-${context}`)
}

// ============================================================================
// CATEGORY SELECTORS
// ============================================================================

/**
 * Get all selectors from a category as a record
 */
export function getSelectorsByCategory(category: SelectorCategory): Record<string, string> {
  return SELECTORS[category] as Record<string, string>
}

/**
 * Get all selectors flattened into a single object
 */
export function getAllSelectors(): Record<string, string> {
  const all: Record<string, string> = {}
  Object.values(SELECTORS).forEach((category) => {
    Object.assign(all, category as Record<string, string>)
  })
  return all
}

// ============================================================================
// ELEMENT TYPE HELPERS
// ============================================================================

/**
 * Get a button element with common interactions
 */
export class ButtonHelper {
  constructor(private locator: Locator) {}

  async click(): Promise<void> {
    await this.locator.click()
  }

  async clickAndWaitForNavigation(): Promise<void> {
    await this.locator.click()
  }

  async isVisible(): Promise<boolean> {
    return await this.locator.isVisible()
  }

  async isEnabled(): Promise<boolean> {
    return await this.locator.isEnabled()
  }

  async isDisabled(): Promise<boolean> {
    return await this.locator.isDisabled()
  }

  async getText(): Promise<string> {
    return (await this.locator.textContent()) ?? ''
  }

  async expectToBeVisible(): Promise<void> {
    await expect(this.locator).toBeVisible()
  }

  async expectToBeEnabled(): Promise<void> {
    await expect(this.locator).toBeEnabled()
  }

  async expectToBeDisabled(): Promise<void> {
    await expect(this.locator).toBeDisabled()
  }
}

/**
 * Get an input element with common interactions
 */
export class InputHelper {
  constructor(private locator: Locator) {}

  async fill(value: string): Promise<void> {
    await this.locator.fill(value)
  }

  async clear(): Promise<void> {
    await this.locator.clear()
  }

  async type(value: string): Promise<void> {
    await this.locator.type(value)
  }

  async getValue(): Promise<string> {
    return await this.locator.inputValue()
  }

  async isVisible(): Promise<boolean> {
    return await this.locator.isVisible()
  }

  async isEnabled(): Promise<boolean> {
    return await this.locator.isEnabled()
  }

  async expectToBeVisible(): Promise<void> {
    await expect(this.locator).toBeVisible()
  }

  async expectToHaveValue(value: string): Promise<void> {
    await expect(this.locator).toHaveValue(value)
  }

  async expectToBeEmpty(): Promise<void> {
    await expect(this.locator).toHaveValue('')
  }
}

/**
 * Get a select/dropdown element with common interactions
 */
export class SelectHelper {
  constructor(private locator: Locator) {}

  async selectOption(option: string): Promise<void> {
    await this.locator.selectOption(option)
  }

  async selectOptionByLabel(label: string): Promise<void> {
    await this.locator.selectOption({ label })
  }

  async selectOptionByValue(value: string): Promise<void> {
    await this.locator.selectOption({ value })
  }

  async getValue(): Promise<string> {
    return await this.locator.inputValue()
  }

  async expectToHaveValue(value: string): Promise<void> {
    await expect(this.locator).toHaveValue(value)
  }
}

/**
 * Get a checkbox element with common interactions
 */
export class CheckboxHelper {
  constructor(private locator: Locator) {}

  async check(): Promise<void> {
    await this.locator.check()
  }

  async uncheck(): Promise<void> {
    await this.locator.uncheck()
  }

  async isChecked(): Promise<boolean> {
    return await this.locator.isChecked()
  }

  async expectToBeChecked(): Promise<void> {
    await expect(this.locator).toBeChecked()
  }

  async expectNotToBeChecked(): Promise<void> {
    await expect(this.locator).not.toBeChecked()
  }
}

/**
 * Get a link element with common interactions
 */
export class LinkHelper {
  constructor(private locator: Locator) {}

  async click(): Promise<void> {
    await this.locator.click()
  }

  async clickAndWaitForNavigation(): Promise<void> {
    await this.locator.click()
  }

  async getHref(): Promise<string> {
    return (await this.locator.getAttribute('href')) ?? ''
  }

  async getText(): Promise<string> {
    return (await this.locator.textContent()) ?? ''
  }

  async expectToHaveHref(href: string | RegExp): Promise<void> {
    await expect(this.locator).toHaveAttribute('href', href)
  }
}

/**
 * Get a tab element with common interactions
 */
export class TabHelper {
  constructor(private locator: Locator) {}

  async click(): Promise<void> {
    await this.locator.click()
  }

  async isVisible(): Promise<boolean> {
    return await this.locator.isVisible()
  }

  async isActive(): Promise<boolean> {
    return await this.locator.evaluate((el) => el.classList.contains('active'))
  }

  async expectToBeActive(): Promise<void> {
    await expect(this.locator).toHaveClass(/active/)
  }
}

/**
 * Get a modal element with common interactions
 */
export class ModalHelper {
  constructor(private locator: Locator) {}

  async isVisible(): Promise<boolean> {
    return await this.locator.isVisible()
  }

  async expectToBeVisible(): Promise<void> {
    await expect(this.locator).toBeVisible()
  }

  async expectNotToBeVisible(): Promise<void> {
    await expect(this.locator).not.toBeVisible()
  }

  async expectToContainText(text: string): Promise<void> {
    await expect(this.locator).toContainText(text)
  }

  getCloseButton(): Locator {
    return this.locator.locator('[data-testid*="btn-modal-close"]')
  }

  async close(): Promise<void> {
    const closeBtn = this.getCloseButton()
    if (await closeBtn.isVisible()) {
      await closeBtn.click()
    }
  }
}

/**
 * Get a card element with common interactions
 */
export class CardHelper {
  constructor(private locator: Locator) {}

  async isVisible(): Promise<boolean> {
    return await this.locator.isVisible()
  }

  async expectToBeVisible(): Promise<void> {
    await expect(this.locator).toBeVisible()
  }

  async expectToContainText(text: string): Promise<void> {
    await expect(this.locator).toContainText(text)
  }

  getChildByTestId(testId: string): Locator {
    return this.locator.locator(`[data-testid="${testId}"]`)
  }

  getButton(action: string): ButtonHelper {
    return new ButtonHelper(this.locator.locator(`[data-testid="btn-${action}"]`))
  }
}

/**
 * Get a list element with common interactions
 */
export class ListHelper {
  constructor(private locator: Locator) {}

  async getItems(): Promise<Locator[]> {
    return await this.locator.locator('[data-testid^="item-"]').all()
  }

  async getItemCount(): Promise<number> {
    return (await this.getItems()).length
  }

  async expectItemCount(count: number): Promise<void> {
    await expect(this.locator.locator('[data-testid^="item-"]')).toHaveCount(count)
  }

  async expectToBeEmpty(): Promise<void> {
    await expect(this.locator.locator('[data-testid^="item-"]')).toHaveCount(0)
  }

  getItemByText(text: string): Locator {
    return this.locator.locator('[data-testid^="item-"]', { hasText: text })
  }
}

/**
 * Get a form element with common interactions
 */
export class FormHelper {
  constructor(private locator: Locator) {}

  async fill(fieldName: string, value: string): Promise<void> {
    await getInput(this.locator, fieldName).fill(value)
  }

  async clear(fieldName: string): Promise<void> {
    await getInput(this.locator, fieldName).clear()
  }

  async submit(): Promise<void> {
    await this.locator.locator('[data-testid^="btn-"][type="submit"]').click()
  }

  getField(fieldName: string): Locator {
    return getInput(this.locator, fieldName)
  }

  getSubmitButton(): Locator {
    return this.locator.locator('[data-testid^="btn-"][type="submit"]')
  }

  async expectToBeValid(): Promise<void> {
    await expect(this.locator).not.toHaveClass(/invalid/)
  }

  async expectToBeInvalid(): Promise<void> {
    await expect(this.locator).toHaveClass(/invalid/)
  }
}

// ============================================================================
// PAGE HELPERS
// ============================================================================

/**
 * Navigate to a page and wait for it to load
 */
export async function goToPage(
  page: Page,
  url: string,
  options?: { waitUntil?: 'load' | 'domcontentloaded' | 'networkidle' | 'commit' }
): Promise<void> {
  await page.goto(url, {
    waitUntil: options?.waitUntil ?? 'networkidle',
  })
}

/**
 * Wait for a page to be fully loaded
 */
export async function waitForPageLoad(
  page: Page,
  timeout: number = 10000
): Promise<void> {
  await page.waitForLoadState('networkidle', { timeout })
}

/**
 * Wait for a specific element to be visible
 */
export async function waitForElement(
  page: Page,
  testId: string,
  timeout: number = 10000
): Promise<Locator> {
  const locator = getByTestId(page, testId)
  await locator.waitFor({ state: 'visible', timeout })
  return locator
}

/**
 * Wait for an element to disappear
 */
export async function waitForElementToDisappear(
  page: Page,
  testId: string,
  timeout: number = 10000
): Promise<void> {
  const locator = getByTestId(page, testId)
  await locator.waitFor({ state: 'hidden', timeout })
}

/**
 * Wait for a URL to match a pattern
 */
export async function waitForUrl(
  page: Page,
  pattern: string | RegExp,
  timeout: number = 10000
): Promise<void> {
  await expect(page).toHaveURL(pattern, { timeout })
}

/**
 * Wait for navigation to complete
 */
export async function waitForNavigation(
  page: Page,
  timeout: number = 10000
): Promise<void> {
  await page.waitForURL('**/*', { timeout })
}

// ============================================================================
// FORM HELPERS
// ============================================================================

/**
 * Fill a form with multiple fields
 */
export async function fillForm(
  page: Page,
  fields: Record<string, string>
): Promise<void> {
  for (const [fieldName, value] of Object.entries(fields)) {
    await getInput(page, fieldName).fill(value)
  }
}

/**
 * Clear a form with multiple fields
 */
export async function clearForm(
  page: Page,
  fieldNames: string[]
): Promise<void> {
  for (const fieldName of fieldNames) {
    await getInput(page, fieldName).clear()
  }
}

/**
 * Submit a form and wait for result
 */
export async function submitForm(
  page: Page,
  formLocator: Locator,
  expectedUrlPattern?: string | RegExp
): Promise<void> {
  const submitBtn = formLocator.locator('[data-testid^="btn-"][type="submit"]')
  await submitBtn.click()

  if (expectedUrlPattern) {
    await waitForUrl(page, expectedUrlPattern)
  }
}

// ============================================================================
// ASSERTION HELPERS
// ============================================================================

/**
 * Assert that an element contains expected text
 */
export async function assertText(
  pageOrLocator: Page | Locator,
  testId: string,
  expected: string | RegExp
): Promise<void> {
  await expect(getByTestId(pageOrLocator, testId)).toContainText(expected)
}

/**
 * Assert that an element is visible
 */
export async function assertVisible(
  pageOrLocator: Page | Locator,
  testId: string
): Promise<void> {
  await expect(getByTestId(pageOrLocator, testId)).toBeVisible()
}

/**
 * Assert that an element is hidden
 */
export async function assertHidden(
  pageOrLocator: Page | Locator,
  testId: string
): Promise<void> {
  await expect(getByTestId(pageOrLocator, testId)).toBeHidden()
}

/**
 * Assert that an element has a specific value
 */
export async function assertValue(
  pageOrLocator: Page | Locator,
  testId: string,
  expected: string
): Promise<void> {
  await expect(getByTestId(pageOrLocator, testId)).toHaveValue(expected)
}

/**
 * Assert that a button is enabled
 */
export async function assertButtonEnabled(
  pageOrLocator: Page | Locator,
  action: string,
  target?: string
): Promise<void> {
  await expect(getButton(pageOrLocator, action, target)).toBeEnabled()
}

/**
 * Assert that a button is disabled
 */
export async function assertButtonDisabled(
  pageOrLocator: Page | Locator,
  action: string,
  target?: string
): Promise<void> {
  await expect(getButton(pageOrLocator, action, target)).toBeDisabled()
}

// ============================================================================
// WAIT HELPERS
// ============================================================================

/**
 * Wait for a specific time (use sparingly)
 */
export async function wait(ms: number): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * Wait for loading to complete
 */
export async function waitForLoadingToComplete(
  page: Page,
  context: string = 'spinner'
): Promise<void> {
  const loading = getLoading(page, context)
  if (await loading.isVisible()) {
    await loading.waitFor({ state: 'hidden' })
  }
}

/**
 * Wait for API response
 */
export async function waitForApiResponse(
  page: Page,
  urlPattern: string | RegExp,
  timeout: number = 10000
): Promise<void> {
  await page.waitForResponse(urlPattern, { timeout })
}

/**
 * Wait for all API responses to settle
 */
export async function waitForAllApiResponses(
  page: Page,
  timeout: number = 5000
): Promise<void> {
  await page.waitForLoadState('networkidle', { timeout })
}

// ============================================================================
// DEBUG HELPERS
// ============================================================================

/**
 * Take a screenshot for debugging
 */
export async function takeScreenshot(
  page: Page,
  name: string,
  options?: { fullPage?: boolean; timeout?: number }
): Promise<void> {
  await page.screenshot({
    path: `test-results/screenshots/${name}-${Date.now()}.png`,
    fullPage: options?.fullPage ?? false,
    timeout: options?.timeout ?? 5000,
  })
}

/**
 * Print page console logs
 */
export async function printConsoleLogs(page: Page): Promise<void> {
  page.on('console', (msg) => {
    console.log(`[CONSOLE ${msg.type()}] ${msg.text()}`)
  })
}

/**
 * Print page errors
 */
export async function printPageErrors(page: Page): Promise<void> {
  page.on('pageerror', (error) => {
    console.error(`[PAGE ERROR] ${error.message}`)
  })
}

// ============================================================================
// LOCATOR FACTORY FUNCTIONS
// ============================================================================

/**
 * Create a button helper for a specific button
 */
export function button(
  pageOrLocator: Page | Locator,
  action: string,
  target?: string
): ButtonHelper {
  return new ButtonHelper(getButton(pageOrLocator, action, target))
}

/**
 * Create an input helper for a specific input
 */
export function input(
  pageOrLocator: Page | Locator,
  fieldName: string
): InputHelper {
  return new InputHelper(getInput(pageOrLocator, fieldName))
}

/**
 * Create a select helper for a specific select
 */
export function select(
  pageOrLocator: Page | Locator,
  fieldName: string
): SelectHelper {
  return new SelectHelper(getByTestId(pageOrLocator, `select-${fieldName}`))
}

/**
 * Create a checkbox helper for a specific checkbox
 */
export function checkbox(
  pageOrLocator: Page | Locator,
  name: string
): CheckboxHelper {
  return new CheckboxHelper(getByTestId(pageOrLocator, `checkbox-${name}`))
}

/**
 * Create a link helper for a specific link
 */
export function link(
  pageOrLocator: Page | Locator,
  destination: string
): LinkHelper {
  return new LinkHelper(getLink(pageOrLocator, destination))
}

/**
 * Create a tab helper for a specific tab
 */
export function tab(
  pageOrLocator: Page | Locator,
  tabName: string
): TabHelper {
  return new TabHelper(getTab(pageOrLocator, tabName))
}

/**
 * Create a modal helper for a specific modal
 */
export function modal(
  pageOrLocator: Page | Locator,
  purpose: string
): ModalHelper {
  return new ModalHelper(getModal(pageOrLocator, purpose))
}

/**
 * Create a card helper for a specific card
 */
export function card(
  pageOrLocator: Page | Locator,
  contentType: string
): CardHelper {
  return new CardHelper(getCard(pageOrLocator, contentType))
}

/**
 * Create a list helper for a specific list
 */
export function list(
  pageOrLocator: Page | Locator,
  contentType: string
): ListHelper {
  return new ListHelper(getList(pageOrLocator, contentType))
}

/**
 * Create a form helper for a specific form
 */
export function form(
  pageOrLocator: Page | Locator,
  formName: string
): FormHelper {
  return new FormHelper(getByTestId(pageOrLocator, `form-${formName}`))
}
