import { Locator, Page, expect } from '@playwright/test'

/**
 * Organization Switcher Page Object
 * Handles organization selection and multi-org functionality.
 */
export class OrgSwitcher {
  readonly page: Page

  // Main trigger button
  readonly trigger: Locator
  readonly currentOrgName: Locator
  readonly dropdownButton: Locator

  // Dropdown container
  readonly dropdown: Locator
  readonly dropdownList: Locator

  // Organization items
  readonly orgItems: Locator

  // Create org button
  readonly createOrgButton: Locator

  constructor(page: Page) {
    this.page = page

    // Main trigger - OrganizationSwitcher uses SelectTrigger with avatar
    // The first button with role="combobox" in the header is the org switcher
    this.trigger = page.locator('header button[role="combobox"]').first()
    this.currentOrgName = this.trigger.locator('.truncate')
    this.dropdownButton = this.trigger

    // Dropdown - Radix Select renders content with role="listbox"
    this.dropdown = page.locator('[role="listbox"]').first()
    this.dropdownList = this.dropdown.locator('[role="group"]')

    // Organization items - uses SelectItem (role="option")
    this.orgItems = this.dropdown.locator('[role="option"]').filter({ 
      has: page.locator('.truncate')
    })

    // Create org button - "Yeni Organizasyon" option
    this.createOrgButton = page.locator('[role="option"]').filter({ hasText: /Yeni Organizasyon/i })
  }

  // ============================================================================
  // Assertions
  // ============================================================================

  /**
   * Assert org switcher trigger is visible
   */
  async expectVisible(): Promise<void> {
    await expect(this.trigger).toBeVisible({ timeout: 10000 })
  }

  /**
   * Assert org switcher is hidden
   */
  async expectHidden(): Promise<void> {
    await expect(this.trigger).toBeHidden({ timeout: 5000 })
  }

  /**
   * Assert dropdown is visible
   */
  async expectDropdownVisible(): Promise<void> {
    await expect(this.dropdown).toBeVisible({ timeout: 5000 })
  }

  /**
   * Assert dropdown is hidden
   */
  async expectDropdownHidden(): Promise<void> {
    await expect(this.dropdown).toBeHidden({ timeout: 5000 })
  }

  /**
   * Assert current organization name matches expected
   */
  async expectCurrentOrg(name: string): Promise<void> {
    await expect(this.currentOrgName).toContainText(name)
  }

  /**
   * Assert org item count matches expected
   */
  async expectOrgCount(count: number): Promise<void> {
    await expect(this.orgItems).toHaveCount(count)
  }

  /**
   * Assert specific org is in the list
   */
  async expectOrgInList(orgName: string): Promise<void> {
    await expect(this.orgItems.filter({ hasText: orgName })).toHaveCount(1)
  }

  /**
   * Assert create org button is visible
   */
  async expectCreateOrgButtonVisible(): Promise<void> {
    await expect(this.createOrgButton).toBeVisible()
  }

  /**
   * Assert create org button is hidden
   */
  async expectCreateOrgButtonHidden(): Promise<void> {
    await expect(this.createOrgButton).toBeHidden()
  }

  // ============================================================================
  // Actions
  // ============================================================================

  /**
   * Click on the org switcher trigger
   */
  async click(): Promise<void> {
    await this.dropdownButton.click()
  }

  /**
   * Open dropdown by clicking trigger (only if not already open)
   */
  async openDropdown(): Promise<void> {
    const isOpen = await this.dropdown.isVisible()
    if (!isOpen) {
      await this.click()
    }
    await this.expectDropdownVisible()
  }

  /**
   * Close dropdown by pressing Escape (Radix Select intercepts clicks when open)
   */
  async closeDropdown(): Promise<void> {
    await this.page.keyboard.press('Escape')
    await this.expectDropdownHidden()
  }

  /**
   * Select organization by name
   */
  async selectOrg(orgName: string): Promise<void> {
    await this.openDropdown()
    const orgItem = this.orgItems.filter({ hasText: orgName })
    await expect(orgItem).toBeVisible({ timeout: 5000 })
    await orgItem.click()
    await this.expectDropdownHidden()
  }

  /**
   * Hover over an org item
   */
  async hoverOrgItem(orgName: string): Promise<void> {
    await this.openDropdown()
    const orgItem = this.orgItems.filter({ hasText: orgName })
    await orgItem.hover()
  }

  /**
   * Click create organization button
   */
  async clickCreateOrg(): Promise<void> {
    await this.openDropdown()
    await this.createOrgButton.click()
  }

  // ============================================================================
  // Utility Methods
  // ============================================================================

  /**
   * Get current organization name
   */
  async getCurrentOrgName(): Promise<string> {
    return await this.currentOrgName.textContent() ?? ''
  }

  /**
   * Get list of all organization names
   */
  async getOrgNames(): Promise<string[]> {
    await this.openDropdown()
    const items = await this.orgItems.all()
    const names = await Promise.all(items.map((item) => item.textContent()))
    return names.filter((n) => n !== null) as string[]
  }

  /**
   * Check if specific org is visible
   */
  async isOrgVisible(orgName: string): Promise<boolean> {
    return await this.orgItems.filter({ hasText: orgName }).isVisible()
  }

  /**
   * Check if switcher is visible
   */
  async isVisible(): Promise<boolean> {
    return await this.trigger.isVisible()
  }

  /**
   * Get org item count
   */
  async getOrgCount(): Promise<number> {
    return await this.orgItems.count()
  }
}
