import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import PlaygroundPreview from '../PlaygroundPreview'
vi.mock('@widget/widgetApp', () => {
  return {
    WidgetApp: (props: any) => {
      ;(globalThis as any).__widgetProps = props
      return <div data-testid="widget-app" />
    }
  }
})

describe('PlaygroundPreview', () => {
  it('renders preview toolbar and example domain', () => {
    render(
      <PlaygroundPreview
        id="123"
        themeColor="#a78bfa"
        chatHeaderColor="#3b82f6"
        chatHeaderTextColor="#ffffff"
        botMessageColor="#fcfcfd"
        botMessageTextColor="#030303"
        userMessageColor="#2e408a"
        userMessageTextColor="#ffffff"
        chatFontFamily="Inter, sans-serif"
        position="bottom-right"
        botDisplayName="Destek"
        botIcon=""
        chatBackgroundColor="#FFF5E6"
        welcomeMessage="Merhaba"
        previewOpen={false}
        sessionId="sid"
        suggestionsEnabled={false}
        suggestedQuestions={[]}
      />
    )
    expect(screen.getByText('example.com')).toBeInTheDocument()
  })

  it('passes suggestions and autoOpen props', () => {
    render(
      <PlaygroundPreview
        id="preview"
        themeColor="#a78bfa"
        chatHeaderColor="#3b82f6"
        chatHeaderTextColor="#ffffff"
        botMessageColor="#fcfcfd"
        botMessageTextColor="#030303"
        userMessageColor="#2e408a"
        userMessageTextColor="#ffffff"
        chatFontFamily="Inter, sans-serif"
        position="bottom-right"
        botDisplayName="Destek"
        botIcon=""
        chatBackgroundColor="#FFF5E6"
        welcomeMessage="Merhaba"
        previewOpen={true}
        sessionId="sid"
        suggestionsEnabled={true}
        suggestedQuestions={["A?","B?"]}
      />
    )
    const props = (globalThis as any).__widgetProps
    expect(props.autoOpen).toBe(true)
    expect(props.suggestions).toEqual(["A?","B?"])
  })

  it('disables suggestions when suggestionsEnabled=false', () => {
    render(
      <PlaygroundPreview
        id="preview"
        themeColor="#a78bfa"
        chatHeaderColor="#3b82f6"
        chatHeaderTextColor="#ffffff"
        botMessageColor="#fcfcfd"
        botMessageTextColor="#030303"
        userMessageColor="#2e408a"
        userMessageTextColor="#ffffff"
        chatFontFamily="Inter, sans-serif"
        position="bottom-right"
        botDisplayName="Destek"
        botIcon=""
        chatBackgroundColor="#FFF5E6"
        welcomeMessage="Merhaba"
        previewOpen={false}
        sessionId="sid"
        suggestionsEnabled={false}
        suggestedQuestions={["A?", "B?"]}
      />
    )
    const props = (globalThis as any).__widgetProps
    expect(props.autoOpen).toBe(false)
    expect(props.suggestions).toEqual([])
  })
})
