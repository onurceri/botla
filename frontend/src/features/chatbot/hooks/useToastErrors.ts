import { useToast } from '@/components/ui/toast'

export function useToastErrors() {
  const { toast } = useToast()
  return {
    success: (message: string) => toast(message, 'success'),
    error: (message: string) => toast(message, 'error'),
    info: (message: string) => toast(message, 'info'),
  }
}
