import { useState, useRef, useEffect } from 'react'
import { RgbaColorPicker } from 'react-colorful'
import * as Popover from '@radix-ui/react-popover'
import { parseColor, formatRgba, type RgbaColor } from '@/lib/color-utils'

interface ColorPickerProps {
  value: string
  onChange: (value: string) => void
  id?: string
  className?: string
  label?: string
}

export function ColorPicker({ value, onChange, id, className, label }: ColorPickerProps) {
  const [internalValue, setInternalValue] = useState(value)
  const [isOpen, setIsOpen] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  // Parse the current color value
  const rgbaColor = parseColor(value)

  // Sync internal value with external value
  useEffect(() => {
    setInternalValue(value)
  }, [value])

  const handleColorChange = (color: RgbaColor) => {
    const formatted = formatRgba(color)
    setInternalValue(formatted)
    onChange(formatted)
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value
    setInternalValue(newValue)
    // Only update if valid color
    const parsed = parseColor(newValue)
    if (parsed.r !== 0 || parsed.g !== 0 || parsed.b !== 0 || newValue.startsWith('#') || newValue.startsWith('rgb')) {
      onChange(newValue)
    }
  }

  const handleInputBlur = () => {
    // On blur, normalize the value
    const parsed = parseColor(internalValue)
    const formatted = formatRgba(parsed)
    setInternalValue(formatted)
    onChange(formatted)
  }

  return (
    <div className={`flex items-center ${className || ''}`}>
      <Popover.Root open={isOpen} onOpenChange={setIsOpen}>
        <Popover.Trigger asChild>
          <button
            type="button"
            aria-label={label || 'Renk seç'}
            className="group relative w-14 h-10 rounded-full overflow-hidden border-2 border-slate-200/60 bg-white shrink-0 cursor-pointer transition-all hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5 focus:outline-none focus:ring-4 focus:ring-primary/10 active:scale-95"
          >
            <div 
              className="absolute inset-1 rounded-full shadow-inner transition-transform group-hover:scale-[1.02]"
              style={{
                backgroundColor: value,
              }}
            />
            {/* Checkerboard pattern for transparency */}
            <div
              className="absolute inset-1 rounded-full -z-10"
              style={{
                backgroundImage: `
                  linear-gradient(45deg, #eee 25%, transparent 25%),
                  linear-gradient(-45deg, #eee 25%, transparent 25%),
                  linear-gradient(45deg, transparent 75%, #eee 75%),
                  linear-gradient(-45deg, transparent 75%, #eee 75%)
                `,
                backgroundSize: '8px 8px',
                backgroundPosition: '0 0, 0 4px, 4px -4px, -4px 0px',
              }}
            />
          </button>
        </Popover.Trigger>
        <Popover.Portal>
          <Popover.Content
            className="z-50 rounded-2xl bg-white p-4 shadow-[0_10px_40px_rgba(0,0,0,0.1)] border border-slate-200/60 animate-in fade-in-0 zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2"
            sideOffset={12}
            align="start"
          >
            <RgbaColorPicker 
              color={rgbaColor} 
              onChange={handleColorChange}
            />
            <Popover.Arrow className="fill-white" />
          </Popover.Content>
        </Popover.Portal>
      </Popover.Root>
    </div>
  )
}
