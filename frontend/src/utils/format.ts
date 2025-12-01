export const formatDate = (d: Date) => d.toLocaleDateString()
export const formatNumber = (n: number) => new Intl.NumberFormat().format(n)
