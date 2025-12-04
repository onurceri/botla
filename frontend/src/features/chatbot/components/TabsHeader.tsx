import { TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Settings, Database, Play, Code, MessageSquare } from 'lucide-react'

export default function TabsHeader() {
  return (
    <div className="overflow-x-auto pb-2 -mx-4 px-4 md:mx-0 md:px-0 md:pb-0 scrollbar-hide">
      <TabsList className="h-auto w-max flex-nowrap justify-start gap-2 md:w-auto md:flex-wrap">
        <TabsTrigger value="overview" className="gap-2 whitespace-nowrap">
          <Settings className="w-4 h-4" /> Genel
        </TabsTrigger>
        <TabsTrigger value="sources" className="gap-2 whitespace-nowrap">
          <Database className="w-4 h-4" /> Veri Kaynakları
        </TabsTrigger>
        <TabsTrigger value="playground" className="gap-2 whitespace-nowrap">
          <Play className="w-4 h-4" /> Playground
        </TabsTrigger>
        <TabsTrigger value="connect" className="gap-2 whitespace-nowrap">
          <Code className="w-4 h-4" /> Entegrasyon
        </TabsTrigger>
        <TabsTrigger value="suggestions" className="gap-2 whitespace-nowrap">
          <MessageSquare className="w-4 h-4" /> Örnek Sorular
        </TabsTrigger>
      </TabsList>
    </div>
  )
}

