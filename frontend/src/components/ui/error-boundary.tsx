import { Component, ErrorInfo, ReactNode } from 'react'
import { AlertCircle } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  }

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error:', error, errorInfo)
  }

  public render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <Alert variant="destructive" className="my-4">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Bir hata oluştu</AlertTitle>
          <AlertDescription className="mt-2 flex flex-col gap-2">
            <p>Bu bileşen görüntülenirken bir sorun oluştu.</p>
            <div className="text-xs font-mono bg-background/10 p-2 rounded">
              {this.state.error?.message}
            </div>
            <Button 
              variant="outline" 
              size="sm" 
              className="w-fit mt-2"
              onClick={() => this.setState({ hasError: false, error: null })}
            >
              Tekrar Dene
            </Button>
          </AlertDescription>
        </Alert>
      )
    }

    return this.props.children
  }
}
