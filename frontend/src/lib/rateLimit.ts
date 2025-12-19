
type RateLimitState = {
    limit: number | null
    remaining: number | null
}

class RateLimitStore {
    private state: RateLimitState = {
        limit: null,
        remaining: null,
    }
    private listeners: Set<(state: RateLimitState) => void> = new Set()

    getState() {
        return this.state
    }

    setState(newState: Partial<RateLimitState>) {
        this.state = { ...this.state, ...newState }
        this.notify()
    }

    updateFromHeaders(headers: Record<string, any>) {
        const limit = headers['x-ratelimit-limit']
        const remaining = headers['x-ratelimit-remaining']

        if (limit || remaining) {
            this.setState({
                limit: limit ? parseInt(limit, 10) : this.state.limit,
                remaining: remaining ? parseInt(remaining, 10) : this.state.remaining,
            })
        }
    }

    subscribe(listener: (state: RateLimitState) => void) {
        this.listeners.add(listener)
        return () => {
            this.listeners.delete(listener)
        }
    }

    private notify() {
        this.listeners.forEach((listener) => listener(this.state))
    }
}

export const rateLimitStore = new RateLimitStore()
