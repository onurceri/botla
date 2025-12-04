import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

type Props = {
  name: string
  setName: (v: string) => void
  systemPrompt: string
  setSystemPrompt: (v: string) => void
}

export default function OverviewPanel({ name, setName, systemPrompt, setSystemPrompt }: Props) {
  return (
    <div className="grid gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Kimlik</CardTitle>
          <CardDescription>Bot ismi ve sistem mesajı.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">İsim</label>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">System Prompt</label>
            <textarea 
              className="flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              value={systemPrompt}
              onChange={(e) => setSystemPrompt(e.target.value)}
              placeholder="Sen yardımcı bir asistansın..."
            />
            <div className="flex justify-end text-xs text-muted-foreground">{systemPrompt.length} karakter</div>
          </div>
        </CardContent>
      </Card>

    </div>
  )
}
