export const decodeJwt = (token: string) => {
  try {
    const payload = token.split('.')[1]
    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64 + '==='.slice((base64.length + 3) % 4)
    const json = atob(padded)
    return JSON.parse(json)
  } catch {
    return null
  }
}

export const getJwtExp = (token: string): number | null => {
  const payload = decodeJwt(token)
  return payload?.exp ? Number(payload.exp) : null
}

export const isJwtExpired = (token: string): boolean => {
  const exp = getJwtExp(token)
  if (!exp) return false
  return Date.now() >= exp * 1000
}
