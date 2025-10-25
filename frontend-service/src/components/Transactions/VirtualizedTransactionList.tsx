import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { FixedSizeList as List } from 'react-window'
import { fetchTransactions } from '../../api'
import type { Transaction, TransactionFilters } from '../../types'
import { formatCurrency, formatDate } from '../../utils/helpers'

interface VirtualizedTransactionListProps {
  token: string
  filters: TransactionFilters
  onTransactionClick?: (transaction: Transaction) => void
}

interface TransactionItemProps {
  index: number
  style: React.CSSProperties
  data: {
    transactions: Transaction[]
    onTransactionClick?: (transaction: Transaction) => void
  }
}

const TransactionItem: React.FC<TransactionItemProps> = ({ index, style, data }) => {
  const { transactions, onTransactionClick } = data
  const transaction = transactions[index]

  if (!transaction) {
    return <div style={style}>Loading...</div>
  }

  const isExpense = transaction.operation_type === 'expense'
  const amountColor = isExpense ? 'text-red-500' : 'text-green-500'
  const amountPrefix = isExpense ? '-' : '+'

  return (
    <div
      style={style}
      className="transaction-item"
      onClick={() => onTransactionClick?.(transaction)}
    >
      <div className="transaction-content">
        <div className="transaction-main">
          <div className="transaction-info">
            <div className="transaction-category">
              {transaction.category_name || 'Без категории'}
              {transaction.subcategory_name && (
                <span className="transaction-subcategory">
                  → {transaction.subcategory_name}
                </span>
              )}
            </div>
            <div className="transaction-meta">
              <span className="transaction-date">
                {formatDate(transaction.timestamp)}
              </span>
              {transaction.username && (
                <span className="transaction-user">
                  {transaction.username}
                </span>
              )}
              {transaction.is_shared && (
                <span className="transaction-shared">👥</span>
              )}
            </div>
          </div>
          <div className={`transaction-amount ${amountColor}`}>
            {amountPrefix}{formatCurrency(transaction.amount_cents)}
          </div>
        </div>
      </div>
    </div>
  )
}

export const VirtualizedTransactionList: React.FC<VirtualizedTransactionListProps> = ({
  token,
  filters,
  onTransactionClick
}) => {
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [nextCursor, setNextCursor] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  // Memoized item data to prevent unnecessary re-renders
  const itemData = useMemo(() => ({
    transactions,
    onTransactionClick
  }), [transactions, onTransactionClick])

  // Load transactions with caching
  const loadTransactions = useCallback(async (cursor?: string, append = false) => {
    if (loading) return

    setLoading(true)
    setError(null)

    try {
      const response = await fetchTransactions(token, {
        ...filters,
        cursor,
        limit: 20
      })

      if (append) {
        setTransactions(prev => [...prev, ...response.transactions])
      } else {
        setTransactions(response.transactions)
      }

      setHasMore(response.pagination.has_more)
      setNextCursor(response.pagination.next_cursor || null)
    } catch (err) {
      setError('Ошибка загрузки транзакций')
      console.error('Failed to load transactions:', err)
    } finally {
      setLoading(false)
    }
  }, [token, filters, loading])

  // Load more transactions (infinite scroll)
  const loadMore = useCallback(() => {
    if (hasMore && nextCursor && !loading) {
      loadTransactions(nextCursor, true)
    }
  }, [hasMore, nextCursor, loading, loadTransactions])

  // Load initial data
  useEffect(() => {
    loadTransactions()
  }, [loadTransactions])

  // Reset when filters change
  useEffect(() => {
    setTransactions([])
    setHasMore(true)
    setNextCursor(null)
    loadTransactions()
  }, [filters.operation_type, filters.category_id, filters.subcategory_id, filters.start_date, filters.end_date])

  // Handle scroll to bottom for infinite scroll
  const handleItemsRendered = useCallback(({ visibleStopIndex }: { visibleStopIndex: number }) => {
    // Load more when user scrolls near the end
    if (visibleStopIndex >= transactions.length - 5 && hasMore && !loading) {
      loadMore()
    }
  }, [transactions.length, hasMore, loading, loadMore])

  // Memoized list height calculation
  const listHeight = useMemo(() => {
    const itemHeight = 80 // Height of each transaction item
    const maxHeight = 400 // Maximum height for the list
    const calculatedHeight = Math.min(transactions.length * itemHeight, maxHeight)
    return Math.max(calculatedHeight, 200) // Minimum height
  }, [transactions.length])

  if (error) {
    return (
      <div className="transaction-list-error">
        <p>{error}</p>
        <button 
          onClick={() => loadTransactions()}
          className="retry-button"
        >
          Попробовать снова
        </button>
      </div>
    )
  }

  if (transactions.length === 0 && !loading) {
    return (
      <div className="transaction-list-empty">
        <p>Транзакции не найдены</p>
      </div>
    )
  }

  return (
    <div className="virtualized-transaction-list">
      <List
        height={listHeight}
        itemCount={transactions.length}
        itemSize={80}
        itemData={itemData}
        onItemsRendered={handleItemsRendered}
        overscanCount={5} // Render 5 extra items for smooth scrolling
      >
        {TransactionItem}
      </List>
      
      {loading && (
        <div className="transaction-list-loading">
          <div className="loading-spinner"></div>
          <span>Загрузка...</span>
        </div>
      )}
      
      {!hasMore && transactions.length > 0 && (
        <div className="transaction-list-end">
          <span>Все транзакции загружены</span>
        </div>
      )}
    </div>
  )
}
