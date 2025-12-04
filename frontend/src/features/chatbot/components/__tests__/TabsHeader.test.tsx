import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Tabs } from '@/components/ui/tabs'
import TabsHeader from '../TabsHeader'

describe('TabsHeader', () => {
  it('renders all tab triggers', () => {
    render(
      <Tabs value="overview" onValueChange={() => {}}>
        <TabsHeader />
      </Tabs>
    )
    expect(screen.getByRole('tab', { name: /Genel/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Veri Kaynakları/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Playground/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Entegrasyon/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Örnek Sorular/i })).toBeInTheDocument()
  })
})

