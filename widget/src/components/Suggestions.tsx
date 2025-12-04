export function Suggestions({ items, disabled, onPick }: { items: string[]; disabled: boolean; onPick: (q: string) => void }) {
  if (!items || items.length === 0) return null
  return (
    <div className="cbw-suggestions" aria-label="Önerilen sorular">
      <div className="cbw-suggestions-header">
        <span>✨ ÖRNEK SORULAR</span>
      </div>
      {items.map((s, i) => (
        <button
          key={i}
          className="cbw-suggestion"
          onClick={() => onPick(s)}
          disabled={disabled}
          aria-label={s}
        >{s}</button>
      ))}
    </div>
  )
}
