import { test, expect } from '@playwright/test'

test('Chunk Inspector Flow: Login -> Create Bot -> Add Source -> Inspect Chunks', async ({ page }) => {
    const isReal = !!process.env.E2E_API_BASE

    if (!isReal) {
        // Generic fallback for unhandled GETs (add first so it's checked last)
        await page.route('**/api/v1/**', async r => {
            if (r.request().method() === 'GET' && !r.request().url().includes('chatbots') && !r.request().url().includes('auth')) {
                await r.fulfill({ status: 200, body: JSON.stringify([]) })
            } else {
                await r.continue()
            }
        })

        // Stub auth and basic endpoints
        await page.route('**/api/v1/auth/register', async r => r.fulfill({ status: 200, body: JSON.stringify({ ok: true }) }))
        await page.route('**/api/v1/auth/login', async r => r.fulfill({ status: 200, body: JSON.stringify({ token: 't', refresh_token: 'rt' }) }))
        await page.route('**/api/v1/auth/me', async r => r.fulfill({ status: 200, body: JSON.stringify({ plan: 'pro', id: 'user-1' }) }))
        await page.route('**/api/v1/analytics', async r => r.fulfill({ status: 200, body: JSON.stringify([]) }))

        // Stub orgs/workspaces
        await page.route('**/api/v1/organizations', async r => r.fulfill({
            status: 200,
            body: JSON.stringify([{
                id: 'org-1',
                name: 'Test Org',
                slug: 'test-org',
                owner_id: 'user-1',
                plan_id: 'pro',
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString()
            }])
        }))
        await page.route('**/workspaces', async r => r.fulfill({
            status: 200,
            body: JSON.stringify([{
                id: 'ws-1',
                organization_id: 'org-1',
                name: 'Test Workspace',
                slug: 'test-ws',
                created_at: new Date().toISOString()
            }])
        }))

        // Stub chatbots list
        await page.route('**/api/v1/chatbots', async r => {
            if (r.request().method() === 'GET') {
                await r.fulfill({ status: 200, body: JSON.stringify([]) })
            } else {
                await r.fulfill({ status: 201, body: JSON.stringify({ id: 'bot-1' }) })
            }
        })
        await page.route('**/api/v1/chatbots/bot-1', async r => r.fulfill({ status: 200, body: JSON.stringify({ id: 'bot-1', name: 'Test Bot' }) }))

        // Stub sources
        await page.route('**/api/v1/chatbots/bot-1/sources', async r => {
            if (r.request().method() === 'POST') {
                await r.fulfill({ status: 201, body: JSON.stringify({ id: 'src-1' }) })
            } else {
                await r.fulfill({
                    status: 200, body: JSON.stringify([{
                        id: 'src-1',
                        source_type: 'text',
                        original_filename: 'test-source.txt',
                        status: 'completed',
                        chunk_count: 5,
                        created_at: new Date().toISOString()
                    }])
                })
            }
        })

        // Stub chunks endpoint
        await page.route('**/api/v1/sources/src-1/chunks*', async r => {
            await r.fulfill({
                status: 200,
                body: JSON.stringify({
                    chunks: Array.from({ length: 5 }, (_, i) => ({
                        id: `c-${i}`,
                        score: 0.9,
                        payload: {
                            source_id: 'src-1',
                            original_text: `Chunk content ${i} for inspection.`,
                            chunk_index: i,
                            created_at: new Date().toISOString()
                        }
                    })),
                    next_cursor: null
                })
            })
        })
    }

    // 1. Login
    const email = `test-${Date.now()}@example.com`
    await page.goto('/register')
    await page.getByLabel('Ad Soyad').fill('Test User')
    await page.getByLabel('Email').fill(email)
    await page.getByLabel('Şifre').fill('password')
    await page.getByRole('button', { name: 'Kayıt Ol' }).click()
    await expect(page).toHaveURL(/.*onboarding|.*login/, { timeout: 10000 })

    if (page.url().includes('onboarding')) {
        await page.goto('/login')
    }

    await page.getByLabel('Email').fill(email)
    await page.getByLabel('Şifre').fill('password')
    await page.getByRole('button', { name: 'Giriş Yap' }).click()

    await expect(page).toHaveURL(/.*dashboard|.*\//, { timeout: 10000 })

    // 2. Create Bot
    const createBtn = page.getByRole('button', { name: /Yeni Chatbot|Yeni Oluştur/ }).first()
    await createBtn.click()
    await page.getByPlaceholder('Örn: Müşteri Temsilcisi').fill('Test Bot')
    await page.getByRole('button', { name: 'Oluştur' }).click()
    await expect(page).toHaveURL(/\/chatbots\/[a-zA-Z0-9_-]+$/, { timeout: 10000 })

    // 3. Go to Sources
    await page.getByRole('link', { name: 'Kaynaklar' }).click()

    // 4. Verify List & Open Inspector
    // If mocked, the list should show 'src-1' immediately.
    // We need to wait for the list to load.
    await expect(page.getByRole('heading', { name: 'test-source.txt' })).toBeVisible()

    // Find "View Chunks" button (Desktop or Mobile).
    const inspectBtn = page.getByRole('button', { name: 'Parçaları İncele' }).first()
    await inspectBtn.click()

    // 5. Verify Inspector Content
    await expect(page.getByText('Kaynak Parçaları')).toBeVisible()
    await expect(page.getByText('Chunk content 0 for inspection.')).toBeVisible()

    // 6. Test Search
    const searchInput = page.getByPlaceholder('Yüklenen parçalarda ara...')
    await searchInput.fill('Chunk content 2')
    await expect(page.getByText('Chunk content 2 for inspection.')).toBeVisible()
    await expect(page.getByText('Chunk content 0 for inspection.')).not.toBeVisible()
})
