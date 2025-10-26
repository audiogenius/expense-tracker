import React, { useState, useEffect } from 'react'
import { fetchTransactions } from '../../api'
import type { Transaction, TransactionFilters } from '../../types'
import { formatCurrency, formatDate } from '../../utils/helpers'

type RecentTransactionsProps = {
  token: string
  onViewAll: () => void
  refreshTrigger?: number
}

export const RecentTransactions: React.FC<RecentTransactionsProps> = ({ token, onViewAll, refreshTrigger }) => {
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [filters, setFilters] = useState<TransactionFilters>({
    operation_type: 'both',
    limit: 10
  })

  useEffect(() => {
    loadTransactions()
  }, [filters, refreshTrigger])

  const loadTransactions = async () => {
    try {
      setLoading(true)
      const response = await fetchTransactions(token, filters)
      setTransactions(response.transactions)
    } catch (error) {
      console.error('Failed to load transactions:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleFilterChange = (operationType: 'expense' | 'income' | 'both') => {
    setFilters(prev => ({ ...prev, operation_type: operationType }))
  }

  const getTransactionColor = (operationType: 'expense' | 'income') => {
    return operationType === 'expense' ? 'var(--error)' : 'var(--success)'
  }

  const getTransactionIcon = (operationType: 'expense' | 'income') => {
    return operationType === 'expense' ? '↓' : '↑'
  }

  if (loading) {
    return (
      <div className="glass-card">
        <h3>Последние операции</h3>
        <div className="loading-skeleton">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="skeleton-item">
              <div className="skeleton-line"></div>
              <div className="skeleton-line short"></div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="glass-card recent-transactions">
      <div className="recent-transactions-header">
        <h3>Последние операции</h3>
        <div className="filter-buttons">
          <button
            className={`filter-btn ${filters.operation_type === 'both' ? 'active' : ''}`}
            onClick={() => handleFilterChange('both')}
          >
            Все
          </button>
          <button
            className={`filter-btn ${filters.operation_type === 'expense' ? 'active' : ''}`}
            onClick={() => handleFilterChange('expense')}
          >
            Расходы
          </button>
          <button
            className={`filter-btn ${filters.operation_type === 'income' ? 'active' : ''}`}
            onClick={() => handleFilterChange('income')}
          >
            Приходы
          </button>
        </div>
      </div>

      <div className="transactions-list">
        {transactions.length === 0 ? (
          <div className="empty-state">
            <p>Нет операций</p>
          </div>
        ) : (
          transactions.map((transaction) => (
            <div key={transaction.id} className="transaction-item">
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
          ))
        )}
      </div>

      <div className="recent-transactions-footer">
        <button className="view-all-btn" onClick={onViewAll}>
          Показать всё
        </button>
      </div>
    </div>
  )
}
