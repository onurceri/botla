import { useMemo, useState } from 'react'

export const usePagination = (pageSize = 10) => {
  const [page, setPage] = useState(1)
  const offset = useMemo(() => (page - 1) * pageSize, [page, pageSize])
  return { page, setPage, pageSize, offset }
}
