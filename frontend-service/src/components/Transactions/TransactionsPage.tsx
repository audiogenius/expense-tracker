import React, { useState, useEffect, useCallback, useRef } from 'react'
import { fetchTransactions, fetchCategories, fetchSubcategories } from '../../api'
import type { Transaction, Category, Subcategory, TransactionFilters } from '../../types'
import { formatCurrency, formatDate } from '../../utils/helpers'

type TransactionsPageProps = {
  token: string
}

const ITEMS_PER_PAGE = 20
const ITEM_HEIGHT = 80
const CONTAINER_HEIGHT = 600

export const TransactionsPage: React.FC<TransactionsPageProps> = ({ token }) => {
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [subcategories, setSubcategories] = useState<Subcategory[]>([])
  const [loading, setLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [currentPage, setCurrentPage] = useState(1)
  const [filters, setFilters] = useState<TransactionFilters>({
    operation_type: 'both',
    limit: ITEMS_PER_PAGE
  })

  const containerRef = useRef<HTMLDivElement>(null)
  const loadingRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    loadInitialData()
  }, [])

  useEffect(() => {
    loadTransactions(true)
  }, [filters])

  const loadInitialData = async () => {
    try {
      const [categoriesData] = await Promise.all([
        fetchCategories()
      ])
      setCategories(categoriesData)
    } catch (error) {
      console.error('Failed to load initial data:', error)
    }
  }

  const loadTransactions = async (reset = false) => {
    if (loading) return

    try {
      setLoading(true)
      const page = reset ? 1 : currentPage
      const response = await fetchTransactions(token, {
        ...filters,
        page,
        limit: ITEMS_PER_PAGE
      })

      if (reset) {
        setTransactions(response.transactions)
        setCurrentPage(1)
      } else {
        setTransactions(prev => [...prev, ...response.transactions])
        setCurrentPage(page)
      }

      setHasMore(response.pagination.has_next)
    } catch (error) {
      console.error('Failed to load transactions:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadMore = useCallback(() => {
    if (!loading && hasMore) {
      setCurrentPage(prev => {
        loadTransactions(false)
        return prev + 1
      })
    }
  }, [loading, hasMore])

  // Intersection Observer for infinite scroll
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loading) {
          loadMore()
        }
      },
      { threshold: 0.1 }
    )

    if (loadingRef.current) {
      observer.observe(loadingRef.current)
    }

    return () => observer.disconnect()
  }, [loadMore, hasMore, loading])

  const handleFilterChange = (key: keyof TransactionFilters, value: any) => {
    setFilters(prev => ({ ...prev, [key]: value }))
  }

  const handleCategoryChange = async (categoryId: number | null) => {
    setFilters(prev => ({ ...prev, category_id: categoryId, subcategory_id: undefined }))
    
    if (categoryId) {
      try {
        const data = await fetchSubcategories(token, categoryId)
        setSubcategories(data)
      } catch (error) {
        console.error('Failed to load subcategories:', error)
      }
    } else {
      setSubcategories([])
    }
  }

  const getTransactionColor = (operationType: 'expense' | 'income') => {
    return operationType === 'expense' ? 'var(--error)' : 'var(--success)'
  }

  const getTransactionIcon = (operationType: 'expense' | 'income') => {
    return operationType === 'expense' ? '↓' : '↑'
  }

  const renderTransaction = (transaction: Transaction, index: number) => (
    <div key={transaction.id} className="transaction-item" style={{ height: ITEM_HEIGHT }}>
      <div className="transaction-icon" style={{ color: getTransactionColor(transaction.operation_type) }}>
        {getTransactionIcon(transaction.operation_type)}
      </div>
      <div className="transaction-details">
        <div className="transaction-amount" style={{ color: getTransactionColor(transaction.operation_type) }}>
          {transaction.operation_type === 'expense' ? '-' : '+'}{formatCurrency(transaction.amount_cents)}
        </div>
        <div className="transaction-info">
          <div className="transaction-category">
            {transaction.subcategory_name || transaction.category_name || 'Без категории'}
          </div>
          <div className="transaction-meta">
            <span className="transaction-user">{transaction.username}</span>
            <span className="transaction-date">{formatDate(transaction.timestamp)}</span>
          </div>
        </div>
      </div>
    </div>
  )

  return (
    <div className="transactions-page">
      <div className="transactions-header">
        <h1>Все транзакции</h1>
        
        <div className="transactions-filters">
          <div className="filter-group">
            <label>Тип операции</label>
            <select
              value={filters.operation_type || 'both'}
              onChange={(e) => handleFilterChange('operation_type', e.target.value)}
              className="filter-select"
            >
              <option value="both">Все</option>
              <option value="expense">Расходы</option>
              <option value="income">Приходы</option>
            </select>
          </div>

          <div className="filter-group">
            <label>Категория</label>
            <select
              value={filters.category_id || ''}
              onChange={(e) => handleCategoryChange(e.target.value ? parseInt(e.target.value) : null)}
              className="filter-select"
            >
              <option value="">Все категории</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </select>
          </div>

          {filters.category_id && subcategories.length > 0 && (
            <div className="filter-group">
              <label>Подкатегория</label>
              <select
                value={filters.subcategory_id || ''}
                onChange={(e) => handleFilterChange('subcategory_id', e.target.value ? parseInt(e.target.value) : undefined)}
                className="filter-select"
              >
                <option value="">Все подкатегории</option>
                {subcategories.map((subcategory) => (
                  <option key={subcategory.id} value={subcategory.id}>
                    {subcategory.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="filter-group">
            <label>Период</label>
            <div className="date-range">
              <input
                type="date"
                value={filters.start_date || ''}
                onChange={(e) => handleFilterChange('start_date', e.target.value || undefined)}
                className="date-input"
              />
              <span>—</span>
              <input
                type="date"
                value={filters.end_date || ''}
                onChange={(e) => handleFilterChange('end_date', e.target.value || undefined)}
                className="date-input"
              />
            </div>
          </div>
        </div>
      </div>

      <div className="transactions-container" ref={containerRef}>
        {transactions.length === 0 && !loading ? (
          <div className="empty-state">
            <p>Нет транзакций</p>
          </div>
        ) : (
          <div className="transactions-list">
            {transactions.map((transaction, index) => renderTransaction(transaction, index))}
          </div>
        )}

        {loading && (
          <div className="loading-indicator">
            <div className="spinner"></div>
            <span>Загрузка...</span>
          </div>
        )}

        <div ref={loadingRef} className="infinite-scroll-trigger" />
      </div>
    </div>
  )
}
