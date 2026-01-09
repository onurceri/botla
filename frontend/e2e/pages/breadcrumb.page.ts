import { Locator, Page, expect } from '@playwright/test'

/**
 * Breadcrumb Navigation Page Object
 * Handles breadcrumb navigation and path verification.
 */
export class Breadcrumb {
  readonly page: Page

  // Main container
  readonly container: Locator

  // Home link
  readonly homeLink: Locator

  // Breadcrumb items
  readonly items: Locator

  // Chevron separators
  readonly separators: Locator

  // Current page indicator
  readonly currentPage: Locator

  // Tooltip (for truncated items)
  readonly tooltip: Locator

  constructor(page: Page) {
    this.page = page

    // Main container - breadcrumb is in header, using a more flexible selector
    this.container = page.locator('header').locator('div').filter({ hasText: 'Botla' }).first()

    // Home link (first span with "Botla")
    this.homeLink = this.container.locator('span').filter({ hasText: 'Botla' })

    // All breadcrumb items (spans in the breadcrumb div)
    this.items = this.container.locator('span')

    // Chevron separators
    this.separators = this.container.locator('svg')

    // Current page (last span - the bold one with current page name)
    this.currentPage = this.container.locator('.text-foreground.font-semibold')

    // Tooltip for truncated items
    this.tooltip = page.locator('[role="tooltip"]')
  }

  // ============================================================================
  // Assertions
  // ============================================================================

  /**
   * Assert breadcrumb container is visible
   */
  async expectVisible(): Promise<void> {
    await expect(this.container).toBeVisible({ timeout: 10000 })
  }

  /**
   * Assert breadcrumb container is hidden
   */
  async expectHidden(): Promise<void> {
    await expect(this.container).toBeHidden({ timeout: 5000 })
  }

  /**
   * Assert item count matches expected
   */
  async expectItemCount(count: number): Promise<void> {
    await expect(this.items).toHaveCount(count)
  }

  /**
   * Assert breadcrumb path matches expected sequence
   */
  async expectPath(path: string[]): Promise<void> {
    const itemTexts = await this.items.allTextContents()
    expect(itemTexts).toEqual(path)
  }

  /**
   * Assert breadcrumb path contains expected items (in order)
   */
  async expectPathContains(pathItems: string[]): Promise<void> {
    const itemTexts = await this.items.allTextContents()
    for (const item of pathItems) {
      expect(itemTexts).toContain(item)
    }
    // Verify order
    let lastIndex = -1
    for (const item of pathItems) {
      const currentIndex = itemTexts.indexOf(item)
      expect(currentIndex).toBeGreaterThan(lastIndex)
      lastIndex = currentIndex
    }
  }

  /**
   * Assert specific item is clickable (not current page)
   */
  async expectItemClickable(index: number): Promise<void> {
    const item = this.items.nth(index)
    await expect(item.locator('a')).toBeAttached()
  }

  /**
   * Assert specific item is not clickable (current page)
   */
  async expectItemNotClickable(index: number): Promise<void> {
    const item = this.items.nth(index)
    await expect(item.locator('a')).not.toBeAttached()
  }

  /**
   * Assert tooltip is visible on hover
   */
  async expectTooltipVisible(): Promise<void> {
    await expect(this.tooltip).toBeVisible({ timeout: 3000 })
  }

  /**
   * Assert tooltip is hidden
   */
  async expectTooltipHidden(): Promise<void> {
    await expect(this.tooltip).toBeHidden({ timeout: 3000 })
  }

  /**
   * Assert home link is visible
   */
  async expectHomeLinkVisible(): Promise<void> {
    await expect(this.homeLink).toBeVisible()
  }

  // ============================================================================
  // Actions
  // ============================================================================

  /**
   * Click on home/breadcrumb link at specified index
   */
  async clickBreadcrumb(index: number): Promise<void> {
    const item = this.items.nth(index)
    const link = item.locator('a')
    await link.click()
  }

  /**
   * Click on home link
   */
  async clickHome(): Promise<void> {
    await this.homeLink.click()
    await expect(this.page).toHaveURL(/\/dashboard$/, { timeout: 10000 })
  }

  /**
   * Click on intermediate breadcrumb item (not current)
   */
  async clickIntermediateItem(index: number): Promise<void> {
    await this.clickBreadcrumb(index)
  }

  /**
   * Hover over breadcrumb item to trigger tooltip
   */
  async hoverItem(index: number): Promise<void> {
    const item = this.items.nth(index)
    await item.hover()
  }

  /**
   * Hover over home link
   */
  async hoverHome(): Promise<void> {
    await this.homeLink.hover()
  }

  /**
   * Navigate through breadcrumb items using keyboard
   */
  async pressTabThroughBreadcrumbs(): Promise<void> {
    for (let i = 0; i < (await this.items.count()); i++) {
      await this.page.keyboard.press('Tab')
    }
  }

  // ============================================================================
  // Utility Methods
  // ============================================================================

  /**
   * Get all breadcrumb item texts
   */
  async getBreadcrumbTexts(): Promise<string[]> {
    // Wait for container to be visible first
    await this.container.waitFor({ state: 'visible', timeout: 5000 })
    return await this.items.allTextContents()
  }

  /**
   * Get breadcrumb item count
   */
  async getItemCount(): Promise<number> {
    return await this.items.count()
  }

  /**
   * Get separator count (should be items - 1)
   */
  async getSeparatorCount(): Promise<number> {
    return await this.separators.count()
  }

  /**
   * Get current page text (last item)
   */
  async getCurrentPageText(): Promise<string> {
    return await this.currentPage.textContent() ?? ''
  }

  /**
   * Check if breadcrumb is visible
   */
  async isVisible(): Promise<boolean> {
    return await this.container.isVisible()
  }

  /**
   * Get tooltip text if visible
   */
  async getTooltipText(): Promise<string> {
    return await this.tooltip.textContent() ?? ''
  }

  /**
   * Check if item at index is truncated (has tooltip)
   */
  async isItemTruncated(index: number): Promise<boolean> {
    const item = this.items.nth(index)
    const boundingBox = await item.boundingBox()
    if (!boundingBox) return false

    // Scroll item into view and check content overflow
    await item.scrollIntoViewIfNeeded()
    const scrollWidth = await item.evaluate((el) => el.scrollWidth)
    const clientWidth = await item.evaluate((el) => el.clientWidth)

    return scrollWidth > clientWidth
  }

  /**
   * Navigate to parent breadcrumb level
   */
  async navigateToParent(): Promise<void> {
    const count = await this.items.count()
    if (count > 1) {
      await this.clickBreadcrumb(count - 2) // Second-to-last item
    }
  }
}
